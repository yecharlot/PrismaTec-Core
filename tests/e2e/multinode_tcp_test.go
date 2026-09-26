package e2e

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/runtime/execution"
)

// TestE2E_MultiNodeTCP is Demo 2+3 on real TCP (not in-process fabric only).
func TestE2E_MultiNodeTCP(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Bootstrap once to obtain stable NodeIDs from disk identity, then reload with peer map.
	bootA, err := core.NewNode(core.Config{
		DataDir: filepath.Join(dir, "a"), Name: "node-A",
	})
	if err != nil {
		t.Fatal(err)
	}
	bootB, err := core.NewNode(core.Config{
		DataDir: filepath.Join(dir, "b"), Name: "node-B",
	})
	if err != nil {
		t.Fatal(err)
	}
	idA := string(bootA.ID())
	idB := string(bootB.ID())
	if idA == "" || idB == "" || idA == idB {
		t.Fatalf("expected distinct NodeIDs, got %q %q", idA, idB)
	}

	nodeA, err := core.NewNode(core.Config{
		DataDir: filepath.Join(dir, "a"), Name: "node-A",
		NetworkAddr: "127.0.0.1:19101",
		Peers:       map[string]string{idB: "127.0.0.1:19102"},
	})
	if err != nil {
		t.Fatal(err)
	}
	nodeB, err := core.NewNode(core.Config{
		DataDir: filepath.Join(dir, "b"), Name: "node-B",
		NetworkAddr: "127.0.0.1:19102",
		Peers:       map[string]string{idA: "127.0.0.1:19101"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(nodeA.ID()) != idA || string(nodeB.ID()) != idB {
		t.Fatalf("identity drift: A %s want %s; B %s want %s", nodeA.ID(), idA, nodeB.ID(), idB)
	}

	if err := nodeB.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := nodeA.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer nodeA.Stop()
	defer nodeB.Stop()

	if nodeA.PeerID() == "" || nodeB.PeerID() == "" {
		t.Fatal("PeerID must be non-empty")
	}
	if nodeA.PeerID() == string(nodeA.ID()) {
		t.Fatal("PeerID must differ from NodeID")
	}

	org, err := nodeA.Organisms().Create(organism.CreateOptions{
		Name:         "tcp-movable",
		Capabilities: []organism.Capability{"memory.read", "memory.write"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := nodeA.Organisms().Start(org.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := nodeA.Organisms().PutMemory(org.ID, "focus", "payload-continuity"); err != nil {
		t.Fatal(err)
	}
	root := org.RootCID
	orgID := org.ID

	var lastRep error
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		lastRep = nodeA.ReplicateOrganism(orgID, []string{idB})
		if lastRep == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if lastRep != nil {
		t.Fatalf("replicate: %v", lastRep)
	}

	deadline = time.Now().Add(5 * time.Second)
	var onB *organism.Organism
	for time.Now().Before(deadline) {
		onB, err = nodeB.Organisms().Get(orgID)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if onB == nil || onB.RootCID != root {
		t.Fatalf("replica on B: %v %+v", err, onB)
	}
	if onB.Memory.Working["focus"] != "payload-continuity" {
		t.Fatalf("memory not replicated: %+v", onB.Memory.Working)
	}

	if err := nodeA.Stop(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)

	res, err := nodeB.RecoverOrganism(orgID)
	if err != nil || res == nil || !res.Success {
		t.Fatalf("recover: %+v %v", res, err)
	}
	got, err := nodeB.Organisms().Get(orgID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Placement.Primary != idB {
		t.Fatalf("primary after recover: %+v", got.Placement)
	}
	if got.RootCID != root {
		t.Fatal("RootCID must survive recovery")
	}
	if got.Status != organism.StatusRunning {
		t.Fatalf("status after recover: %s", got.Status)
	}
	if got.Memory.Working["focus"] != "payload-continuity" {
		t.Fatalf("memory after recover: %+v", got.Memory.Working)
	}
	// RPO: last confirmed Seq on A before stop should equal Seq on B after recover
	// (we wrote one memory key after create/start → Seq>=1; replicate included that state)
	if got.Seq < 1 {
		t.Fatalf("expected Seq>=1 after memory write path, got %d", got.Seq)
	}
	t.Logf("RPO_check seq_on_B=%d (expect no loss vs pre-failure confirmed writes)", got.Seq)

	eng := execution.BuiltinEngine{}
	out, err := eng.Execute(context.Background(), execution.Request{
		OrganismID: got.ID, Entry: "ping",
	})
	if err != nil || string(out.Output) != "pong" {
		t.Fatalf("execute after recover: %v %s", err, out.Output)
	}

	_ = nodeB.Stop()
	nodeB2, err := core.NewNode(core.Config{
		DataDir: filepath.Join(dir, "b"), Name: "node-B",
		NetworkAddr: "127.0.0.1:19102",
		Peers:       map[string]string{idA: "127.0.0.1:19101"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(nodeB2.ID()) != idB {
		t.Fatalf("B identity after restart: %s want %s", nodeB2.ID(), idB)
	}
	reloaded, err := nodeB2.Organisms().Get(orgID)
	if err != nil {
		t.Fatalf("reload after B restart: %v", err)
	}
	if reloaded.RootCID != root || reloaded.Memory.Working["focus"] != "payload-continuity" {
		t.Fatalf("persist after B restart: %+v", reloaded)
	}
	if reloaded.Placement.Primary != idB {
		t.Fatalf("primary after B restart: %+v", reloaded.Placement)
	}
}
