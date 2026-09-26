package network

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestLocalSend(t *testing.T) {
	fab := NewFabric()
	a := NewLocalTransport("node-a", fab)
	b := NewLocalTransport("node-b", fab)
	_ = a.Listen(context.Background())
	_ = b.Listen(context.Background())
	var got atomic.Int32
	b.OnMessage(func(m Message) {
		if m.Type == "ping" {
			got.Add(1)
		}
	})
	if err := a.Send(context.Background(), "node-b", Message{Type: "ping", Payload: []byte("x")}); err != nil {
		t.Fatal(err)
	}
	if got.Load() != 1 {
		t.Fatal("message not delivered")
	}
	if len(a.Peers()) != 1 {
		t.Fatalf("peers %v", a.Peers())
	}
}
