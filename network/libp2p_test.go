package network

import (
	"context"
	"testing"
	"time"
)

func TestLibP2PTransportSend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := NewLibP2PTransport(ctx, "node-A", "/ip4/127.0.0.1/tcp/0", nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewLibP2PTransport(ctx, "node-B", "/ip4/127.0.0.1/tcp/0", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	defer b.Close()
	_ = a.Listen(ctx)
	_ = b.Listen(ctx)

	got := make(chan Message, 1)
	b.OnMessage(func(m Message) {
		if m.Type == "test.ping" {
			got <- m
		}
	})

	// connect A → B using full multiaddr with p2p id
	addrs := b.ListenAddrs()
	if len(addrs) == 0 {
		t.Fatal("no listen addrs")
	}
	if err := a.AddPeer("node-B", addrs[0]); err != nil {
		// may error if connect races; retry
		time.Sleep(200 * time.Millisecond)
		if err2 := a.AddPeer("node-B", addrs[0]); err2 != nil {
			t.Fatal(err2)
		}
	}

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if err := a.Send(ctx, "node-B", Message{Type: "test.ping", Payload: []byte("libp2p-hi")}); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	select {
	case m := <-got:
		if string(m.Payload) != "libp2p-hi" {
			t.Fatalf("%+v", m)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("timeout")
	}
}
