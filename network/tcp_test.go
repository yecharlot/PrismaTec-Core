package network

import (
	"context"
	"testing"
	"time"
)

func TestTCPTransportSend(t *testing.T) {
	a := NewTCPTransport("node-A", DerivePeerID("node-A", "a"), "127.0.0.1:19001")
	b := NewTCPTransport("node-B", DerivePeerID("node-B", "b"), "127.0.0.1:19002")
	a.AddPeer("node-B", "127.0.0.1:19002")
	b.AddPeer("node-A", "127.0.0.1:19001")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := b.Listen(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.Listen(ctx); err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	defer b.Close()

	got := make(chan Message, 1)
	b.OnMessage(func(m Message) {
		if m.Type == "test.ping" {
			got <- m
		}
	})

	// wait for dial
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := a.Send(ctx, "node-B", Message{Type: "test.ping", Payload: []byte("hi")}); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	select {
	case m := <-got:
		if string(m.Payload) != "hi" {
			t.Fatalf("%+v", m)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}
