package e2e

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
)

// TestE2E_MultiNodeTCP is Demo 2+3 on real TCP (not in-process fabric only).
func TestE2E_MultiNodeTCP(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeA, err := core.NewNode(core.Config{
		DataDir:     filepath.Join(dir, "a"),
		Name:        "node-A",
		NetworkAddr: "127.0.0.1:19101",
		Peers:       map[string]string{}, // filled after B id known — use name bridge via second pass
	})
	if err != nil {
		t.Fatal(err)
	}
	nodeB, err := core.NewNode(core.Config{
		DataDir:     filepath.Join(dir, "b"),
		Name:        "node-B",
		NetworkAddr: "127.0.0.1:19102",
		Peers:       map[string]string{},
	})
	if err != nil {
		t.Fatal(err)
	}

	// cross-register by NodeID after creation
	idA := string(nodeA.ID())
	idB := string(nodeB.ID())
	nodeA, _ = core.NewNode(core.Config{
		DataDir: filepath.Join(dir, "a"), Name: "node-A",
		NetworkAddr: "127.0.0.1:19101",
		Peers:       map[string]string{idB: "127.0.0.1:19102"},
	})
	nodeB, _ = core.NewNode(core.Config{
		DataDir: filepath.Join(dir, "b"), Name: "node-B",
		NetworkAddr: "127.0.0.1:19102",
		Peers:       map[string]string{idA: "127.0.0.1:19101"},
	})

	if err := nodeB.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := nodeA.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer nodeA.Stop()
	defer nodeB.Stop()

	org, err := nodeA.Organisms().Create(organism.CreateOptions{
		Name:         "tcp-movable",
		Capabilities: []organism.Capability{"memory.read", "memory.write"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = nodeA.Organisms().Start(org.ID)
	root := org.RootCID

	// wait for TCP link
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		if err := nodeA.ReplicateOrganism(org.ID, []string{string(nodeB.ID())}); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err := nodeA.ReplicateOrganism(org.ID, []string{string(nodeB.ID())}); err != nil {
		t.Fatalf("replicate: %v", err)
	}

	// B should adopt replica
	deadline = time.Now().Add(5 * time.Second)
	var onB *organism.Organism
	for time.Now().Before(deadline) {
		onB, err = nodeB.Organisms().Get(org.ID)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if onB == nil || onB.RootCID != root {
		t.Fatalf("replica on B: %v %+v", err, onB)
	}

	// Demo 3: stop A network → recover on B
	_ = nodeA.Stop()
	time.Sleep(200 * time.Millisecond)

	res, err := nodeB.RecoverOrganism(org.ID)
	if err != nil || res == nil || !res.Success {
		t.Fatalf("recover: %+v %v", res, err)
	}
	got, err := nodeB.Organisms().Get(org.ID)
	if err != nil || got.Placement.Primary != string(nodeB.ID()) {
		t.Fatalf("primary after recover: %+v %v", got, err)
	}
	if got.RootCID != root {
		t.Fatal("RootCID must survive recovery")
	}
	if nodeA.PeerID() == "" || nodeA.PeerID() == string(nodeA.ID()) {
		// PeerID must be distinct format
		if len(nodeB.PeerID()) < 5 {
			t.Fatal("peer id")
		}
	}
}
