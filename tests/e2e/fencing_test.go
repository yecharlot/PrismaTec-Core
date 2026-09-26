package e2e

import (
	"path/filepath"
	"testing"

	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
)

// TestE2E_AntiDualPrimary ensures a stale lower-epoch snapshot cannot overwrite a newer primary.
func TestE2E_AntiDualPrimary(t *testing.T) {
	dir := t.TempDir()
	node, err := core.NewNode(core.Config{DataDir: filepath.Join(dir, "n"), Name: "fence-node"})
	if err != nil {
		t.Fatal(err)
	}
	mgr := node.Organisms()
	org, err := mgr.Create(organism.CreateOptions{
		Name:         "fenced-org",
		Capabilities: []organism.Capability{"memory.read"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if org.Placement.Epoch != 1 {
		t.Fatalf("epoch want 1 got %d", org.Placement.Epoch)
	}
	// Simulate recovered primary at epoch 2
	promoted := org.Snapshot()
	promoted.Placement.Epoch = 2
	promoted.Placement.Primary = string(node.ID())
	promoted.Placement.FencedFrom = "node:old-primary"
	promoted.Status = organism.StatusRunning
	if _, err := mgr.Adopt(promoted, true, "test-promote"); err != nil {
		t.Fatal(err)
	}
	if mgr.EpochOf(org.ID) != 2 {
		t.Fatalf("epoch %d", mgr.EpochOf(org.ID))
	}
	// Stale primary tries to re-assert epoch 1
	stale := org.Snapshot()
	stale.Placement.Epoch = 1
	stale.Placement.Primary = "node:old-primary"
	stale.Status = organism.StatusRunning
	_, err = mgr.Adopt(stale, true, "stale-takeover")
	if err == nil {
		t.Fatal("expected fence rejection of stale epoch")
	}
	// Authoritative state unchanged
	got, err := mgr.Get(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Placement.Epoch != 2 || got.Placement.Primary != string(node.ID()) {
		t.Fatalf("authority broken: %+v", got.Placement)
	}
}
