package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	lp2pnet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/multiformats/go-multiaddr"
)

const ProtocolID = protocol.ID("/prismatec/aip/1.0.0")

// LibP2PTransport implements Transport on libp2p (decentralized P2P fabric).
// PeerID is the libp2p peer.ID — distinct from PrismaTec NodeID.
type LibP2PTransport struct {
	nodeID string
	host   host.Host
	mu     sync.RWMutex
	handlers []Handler
	// logical NodeID → peer.ID
	peers map[string]peer.ID
	// peer.ID → NodeID (reverse)
	byPeer map[peer.ID]string
	online bool
	cancel context.CancelFunc
}

// NewLibP2PTransport builds a host listening on listenMultiaddr (e.g. /ip4/127.0.0.1/tcp/4001).
// priv may be nil (ephemeral key). identitySeed ties key generation for tests if needed.
func NewLibP2PTransport(ctx context.Context, nodeID string, listenMultiaddr string, priv crypto.PrivKey) (*LibP2PTransport, error) {
	if listenMultiaddr == "" {
		listenMultiaddr = "/ip4/0.0.0.0/tcp/0"
	}
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(listenMultiaddr),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
	}
	if priv != nil {
		opts = append(opts, libp2p.Identity(priv))
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("libp2p host: %w", err)
	}
	t := &LibP2PTransport{
		nodeID: nodeID,
		host:   h,
		peers:  map[string]peer.ID{},
		byPeer: map[peer.ID]string{},
	}
	h.SetStreamHandler(ProtocolID, t.handleStream)
	return t, nil
}

func (t *LibP2PTransport) ID() string { return t.nodeID }

func (t *LibP2PTransport) PeerID() string {
	return t.host.ID().String()
}

func (t *LibP2PTransport) Host() host.Host { return t.host }

func (t *LibP2PTransport) Listen(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	t.mu.Lock()
	t.cancel = cancel
	t.online = true
	t.mu.Unlock()
	go func() {
		<-ctx.Done()
		_ = t.Close()
	}()
	return nil
}

func (t *LibP2PTransport) Close() error {
	t.mu.Lock()
	t.online = false
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
	t.mu.Unlock()
	return t.host.Close()
}

func (t *LibP2PTransport) OnMessage(h Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handlers = append(t.handlers, h)
}

// AddPeer registers a remote PrismaTec NodeID bound to a libp2p multiaddr + optional peer id.
// addr formats:
//   - peerID@/ip4/127.0.0.1/tcp/4001
//   - /ip4/127.0.0.1/tcp/4001/p2p/<peerID>
func (t *LibP2PTransport) AddPeer(nodeID, addr string) error {
	pid, ma, err := parsePeerAddr(addr)
	if err != nil {
		return err
	}
	ai := peer.AddrInfo{ID: pid, Addrs: []multiaddr.Multiaddr{}}
	if ma != nil {
		ai.Addrs = append(ai.Addrs, ma)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := t.host.Connect(ctx, ai); err != nil {
		// still register mapping for later dial
		t.mu.Lock()
		t.peers[nodeID] = pid
		t.byPeer[pid] = nodeID
		t.mu.Unlock()
		return fmt.Errorf("libp2p connect %s: %w (registered for retry)", nodeID, err)
	}
	t.mu.Lock()
	t.peers[nodeID] = pid
	t.byPeer[pid] = nodeID
	t.mu.Unlock()
	return nil
}

func (t *LibP2PTransport) Peers() []NodeAddr {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []NodeAddr
	for nid, pid := range t.peers {
		addrs := []string{"p2p://" + pid.String()}
		for _, a := range t.host.Peerstore().Addrs(pid) {
			addrs = append(addrs, a.String())
		}
		out = append(out, NodeAddr{NodeID: nid, Addrs: addrs})
	}
	return out
}

func (t *LibP2PTransport) IsOnline() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.online
}

func (t *LibP2PTransport) Send(ctx context.Context, to string, msg Message) error {
	msg.From = t.nodeID
	msg.To = to
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	t.mu.RLock()
	pid, ok := t.peers[to]
	t.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: %s", ErrPeerOffline, to)
	}
	s, err := t.host.NewStream(ctx, pid, ProtocolID)
	if err != nil {
		return fmt.Errorf("libp2p stream: %w", err)
	}
	defer s.Close()
	return writeJSONStream(s, msg)
}

func (t *LibP2PTransport) Broadcast(ctx context.Context, msg Message) error {
	t.mu.RLock()
	ids := make([]string, 0, len(t.peers))
	for id := range t.peers {
		ids = append(ids, id)
	}
	t.mu.RUnlock()
	var last error
	for _, id := range ids {
		if err := t.Send(ctx, id, msg); err != nil {
			last = err
		}
	}
	return last
}

func (t *LibP2PTransport) handleStream(s lp2pnet.Stream) {
	defer s.Close()
	msg, err := readJSONStream(s)
	if err != nil {
		return
	}
	if msg.From != "" {
		t.mu.Lock()
		t.byPeer[s.Conn().RemotePeer()] = msg.From
		t.peers[msg.From] = s.Conn().RemotePeer()
		t.mu.Unlock()
	}
	t.mu.RLock()
	hs := append([]Handler(nil), t.handlers...)
	t.mu.RUnlock()
	for _, h := range hs {
		h(msg)
	}
}

func writeJSONStream(w io.Writer, msg Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

func readJSONStream(r io.Reader) (Message, error) {
	dec := json.NewDecoder(r)
	var msg Message
	if err := dec.Decode(&msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func parsePeerAddr(addr string) (peer.ID, multiaddr.Multiaddr, error) {
	// peerID@multiaddr
	if i := indexAt(addr); i > 0 {
		pid, err := peer.Decode(addr[:i])
		if err != nil {
			return "", nil, err
		}
		ma, err := multiaddr.NewMultiaddr(addr[i+1:])
		if err != nil {
			return pid, nil, err
		}
		return pid, ma, nil
	}
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return "", nil, err
	}
	pid, err := ma.ValueForProtocol(multiaddr.P_P2P)
	if err != nil {
		return "", ma, fmt.Errorf("multiaddr missing /p2p component: %w", err)
	}
	id, err := peer.Decode(pid)
	return id, ma, err
}

func indexAt(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return i
		}
	}
	return -1
}

// ListenAddrs returns host listen multiaddrs for operators / investors demos.
func (t *LibP2PTransport) ListenAddrs() []string {
	var out []string
	for _, a := range t.host.Addrs() {
		full := a.Encapsulate(multiaddr.StringCast("/p2p/" + t.host.ID().String()))
		out = append(out, full.String())
	}
	return out
}
