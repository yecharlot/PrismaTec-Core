package network

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// TCPTransport is a real multi-process network adapter (manifest Demo 2/3).
// Framing: 4-byte big-endian length + JSON Message envelope.
type TCPTransport struct {
	nodeID  string
	peerID  string
	listen  string // host:port
	mu      sync.RWMutex
	handlers []Handler
	ln      net.Listener
	// peerNodeID → connection
	conns map[string]net.Conn
	// dial targets: nodeID → address
	targets map[string]string
	online  bool
	cancel  context.CancelFunc
}

// NewTCPTransport creates a TCP transport. listenAddr e.g. "127.0.0.1:9001".
func NewTCPTransport(nodeID, peerID, listenAddr string) *TCPTransport {
	return &TCPTransport{
		nodeID:  nodeID,
		peerID:  peerID,
		listen:  listenAddr,
		conns:   map[string]net.Conn{},
		targets: map[string]string{},
	}
}

func (t *TCPTransport) ID() string { return t.nodeID }

func (t *TCPTransport) PeerID() string { return t.peerID }

// AddPeer registers a remote node address for dial (nodeID -> host:port).
func (t *TCPTransport) AddPeer(nodeID, addr string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.targets[nodeID] = addr
}

func (t *TCPTransport) OnMessage(h Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handlers = append(t.handlers, h)
}

func (t *TCPTransport) Listen(ctx context.Context) error {
	ln, err := net.Listen("tcp", t.listen)
	if err != nil {
		return fmt.Errorf("tcp listen %s: %w", t.listen, err)
	}
	ctx, cancel := context.WithCancel(ctx)
	t.mu.Lock()
	t.ln = ln
	t.cancel = cancel
	t.online = true
	t.mu.Unlock()

	go t.acceptLoop(ctx)
	go t.dialLoop(ctx)
	return nil
}

func (t *TCPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.online = false
	if t.cancel != nil {
		t.cancel()
	}
	if t.ln != nil {
		_ = t.ln.Close()
	}
	for id, c := range t.conns {
		_ = c.Close()
		delete(t.conns, id)
	}
	return nil
}

func (t *TCPTransport) Peers() []NodeAddr {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []NodeAddr
	for id, addr := range t.targets {
		if _, ok := t.conns[id]; ok {
			out = append(out, NodeAddr{NodeID: id, Addrs: []string{"tcp://" + addr}})
		}
	}
	// also include connected without target entry
	for id := range t.conns {
		found := false
		for _, o := range out {
			if o.NodeID == id {
				found = true
				break
			}
		}
		if !found {
			out = append(out, NodeAddr{NodeID: id, Addrs: []string{"tcp://" + t.listen}})
		}
	}
	return out
}

func (t *TCPTransport) IsOnline() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.online
}

func (t *TCPTransport) Send(ctx context.Context, to string, msg Message) error {
	msg.From = t.nodeID
	msg.To = to
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	c, err := t.connFor(to)
	if err != nil {
		return err
	}
	return writeMsg(c, msg)
}

func (t *TCPTransport) Broadcast(ctx context.Context, msg Message) error {
	msg.From = t.nodeID
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	t.mu.RLock()
	ids := make([]string, 0, len(t.conns))
	for id := range t.conns {
		ids = append(ids, id)
	}
	t.mu.RUnlock()
	var last error
	for _, id := range ids {
		if id == t.nodeID {
			continue
		}
		if err := t.Send(ctx, id, msg); err != nil {
			last = err
		}
	}
	return last
}

func (t *TCPTransport) connFor(to string) (net.Conn, error) {
	t.mu.RLock()
	c, ok := t.conns[to]
	addr := t.targets[to]
	t.mu.RUnlock()
	if ok {
		return c, nil
	}
	if addr == "" {
		return nil, fmt.Errorf("%w: %s", ErrPeerOffline, to)
	}
	if err := t.dialPeer(to, addr); err != nil {
		return nil, err
	}
	t.mu.RLock()
	c, ok = t.conns[to]
	t.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPeerOffline, to)
	}
	return c, nil
}

func (t *TCPTransport) acceptLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		c, err := t.ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(50 * time.Millisecond)
				continue
			}
		}
		go t.handleConn(ctx, c, true)
	}
}

func (t *TCPTransport) dialLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.mu.RLock()
			targets := make(map[string]string, len(t.targets))
			for k, v := range t.targets {
				targets[k] = v
			}
			t.mu.RUnlock()
			for id, addr := range targets {
				t.mu.RLock()
				_, ok := t.conns[id]
				t.mu.RUnlock()
				if ok {
					continue
				}
				_ = t.dialPeer(id, addr)
			}
		}
	}
}

func (t *TCPTransport) dialPeer(nodeID, addr string) error {
	c, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return err
	}
	// hello
	hello := Message{
		From: t.nodeID, To: nodeID, Type: "net.hello",
		Payload: mustJSON(map[string]string{"node_id": t.nodeID, "peer_id": t.peerID}),
		SentAt:  time.Now().UTC(),
	}
	if err := writeMsg(c, hello); err != nil {
		_ = c.Close()
		return err
	}
	t.registerConn(nodeID, c)
	go t.readLoop(context.Background(), c, nodeID)
	return nil
}

func (t *TCPTransport) handleConn(ctx context.Context, c net.Conn, inbound bool) {
	// first message should be hello to learn remote id
	msg, err := readMsg(c)
	if err != nil {
		_ = c.Close()
		return
	}
	remote := msg.From
	if msg.Type == "net.hello" {
		var m map[string]string
		_ = json.Unmarshal(msg.Payload, &m)
		if m["node_id"] != "" {
			remote = m["node_id"]
		}
		// reply hello
		_ = writeMsg(c, Message{
			From: t.nodeID, To: remote, Type: "net.hello",
			Payload: mustJSON(map[string]string{"node_id": t.nodeID, "peer_id": t.peerID}),
			SentAt:  time.Now().UTC(),
		})
	}
	if remote == "" {
		_ = c.Close()
		return
	}
	t.registerConn(remote, c)
	t.dispatch(msg)
	t.readLoop(ctx, c, remote)
}

func (t *TCPTransport) readLoop(ctx context.Context, c net.Conn, remote string) {
	defer func() {
		t.mu.Lock()
		if cur, ok := t.conns[remote]; ok && cur == c {
			delete(t.conns, remote)
		}
		t.mu.Unlock()
		_ = c.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_ = c.SetReadDeadline(time.Now().Add(60 * time.Second))
		msg, err := readMsg(c)
		if err != nil {
			return
		}
		if msg.From == "" {
			msg.From = remote
		}
		t.dispatch(msg)
	}
}

func (t *TCPTransport) registerConn(nodeID string, c net.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if old, ok := t.conns[nodeID]; ok && old != c {
		_ = old.Close()
	}
	t.conns[nodeID] = c
}

func (t *TCPTransport) dispatch(msg Message) {
	if msg.Type == "net.hello" {
		return
	}
	t.mu.RLock()
	hs := append([]Handler(nil), t.handlers...)
	t.mu.RUnlock()
	for _, h := range hs {
		h(msg)
	}
}

func writeMsg(w io.Writer, msg Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(b)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func readMsg(r io.Reader) (Message, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Message{}, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n > 16*1024*1024 {
		return Message{}, fmt.Errorf("message too large")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return Message{}, err
	}
	var msg Message
	if err := json.Unmarshal(buf, &msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

// ParsePeersEnv parses "nodeID@host:port,nodeID2@host:port".
func ParsePeersEnv(s string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.Index(part, "@"); i > 0 {
			out[part[:i]] = part[i+1:]
		}
	}
	return out
}
