package core

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core/organism"
)

func TestNodeStartStop(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		DataDir: filepath.Join(dir, "node"),
		Name:    "test-node",
	}
	node, err := NewNode(cfg)
	if err != nil {
		t.Fatalf("NewNode: %v", err)
	}
	if node.ID() == "" {
		t.Fatal("expected NodeID")
	}
	if node.Status() != StatusStopped {
		t.Fatalf("expected stopped, got %s", node.Status())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := node.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if node.Status() != StatusRunning {
		t.Fatalf("expected running, got %s", node.Status())
	}

	info := node.Info()
	if info["name"] != "test-node" {
		t.Fatalf("name = %v", info["name"])
	}
	if info["status"] != "running" {
		t.Fatalf("status = %v", info["status"])
	}

	// same identity on recreate
	node2, err := NewNode(cfg)
	if err != nil {
		t.Fatalf("NewNode 2: %v", err)
	}
	if node2.ID() != node.ID() {
		t.Fatalf("identity not persistent: %s vs %s", node2.ID(), node.ID())
	}

	if err := node.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	// allow goroutine to observe cancel
	time.Sleep(20 * time.Millisecond)
}

func TestNodeDoubleStart(t *testing.T) {
	dir := t.TempDir()
	node, err := NewNode(Config{DataDir: dir, Name: "n"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := node.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := node.Start(ctx); err == nil {
		t.Fatal("expected error on double start")
	}
	_ = node.Stop()
}

func TestEventBridgeToPulse(t *testing.T) {
	dir := t.TempDir()
	node, err := NewNode(Config{DataDir: dir, Name: "pulse-node"})
	if err != nil {
		t.Fatal(err)
	}
	org, err := node.Organisms().Create(organism.CreateOptions{Name: "p-agent"})
	if err != nil {
		t.Fatal(err)
	}
	if node.Pulses().Count() < 1 {
		t.Fatalf("expected pulse from organism.created, count=%d", node.Pulses().Count())
	}
	_, err = node.Organisms().Start(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range node.Pulses().Recent(0) {
		if p.Type == "organism.started" && p.Target == org.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("missing organism.started pulse targeted at organism")
	}
}
