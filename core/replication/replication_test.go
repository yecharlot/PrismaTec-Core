package replication

import (
	"context"
	"testing"

	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/network"
)

func TestReplicateAndRecover(t *testing.T) {
	fab := network.NewFabric()
	ta := network.NewLocalTransport("node-A", fab)
	tb := network.NewLocalTransport("node-B", fab)
	_ = ta.Listen(context.Background())
	_ = tb.Listen(context.Background())

	sa := NewService("node-A", ta)
	sb := NewService("node-B", tb)
	tb.OnMessage(sb.HandleMessage)
	ta.OnMessage(sa.HandleMessage)

	org, err := organism.New(organism.CreateOptions{
		Name:         "movable-agent",
		Capabilities: []organism.Capability{"memory.read"},
		NodeID:       "node-A",
	})
	if err != nil {
		t.Fatal(err)
	}
	org.Status = organism.StatusRunning

	snap, err := sa.Replicate(org, []string{"node-B"})
	if err != nil {
		t.Fatal(err)
	}
	if snap.Primary != "node-A" {
		t.Fatalf("primary %s", snap.Primary)
	}

	// B should hold replica
	if sb.GetReplica(org.ID) == nil {
		t.Fatal("replica not on B")
	}

	// simulate A offline
	online := map[string]bool{"node-A": false, "node-B": true}
	sb.SetOnlineFunc(func(id string) bool { return online[id] })
	_ = ta.Close() // remove from fabric peers

	res, err := sb.RecoverIfPrimaryDown(org.ID)
	if err != nil || !res.Success {
		t.Fatalf("recover: %+v %v", res, err)
	}
	if res.NewPrimary != "node-B" || res.OldPrimary != "node-A" {
		t.Fatalf("%+v", res)
	}
	got := sb.GetReplica(org.ID)
	if got.Organism.Status != organism.StatusRunning || got.Primary != "node-B" {
		t.Fatalf("%+v", got)
	}
	if got.Organism.RootCID != org.RootCID {
		t.Fatal("RootCID must be preserved")
	}
}
