// Package e2e runs the Phase 14 end-to-end checklist against real Core primitives.
//
//	go test ./tests/e2e/ -count=1 -v
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/yecharlot/PrismaTec-Core/api/aip"
	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/core/replication"
	"github.com/yecharlot/PrismaTec-Core/network"
	"github.com/yecharlot/PrismaTec-Core/runtime/execution"
	"github.com/yecharlot/PrismaTec-Core/runtime/inference"
)

func TestE2E_ManifestChecklist(t *testing.T) {
	dir := t.TempDir()
	node, err := core.NewNode(core.Config{DataDir: filepath.Join(dir, "node"), Name: "e2e-node"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := node.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer node.Stop()

	mgr := node.Organisms()
	steps := []string{}

	// 1–5 Create + identity + memory + capabilities + start
	org, err := mgr.Create(organism.CreateOptions{
		Name:         "e2e-research-01",
		Capabilities: []organism.Capability{"memory.read", "memory.write", "inference", "execute"},
	})
	if err != nil {
		t.Fatal(err)
	}
	steps = append(steps, "1.create")
	if org.RootCID == "" || org.ID == "" {
		t.Fatal("2.identity: missing RootCID or ID")
	}
	steps = append(steps, "2.identity")
	if len(org.Capabilities) < 2 {
		t.Fatal("4.capabilities missing")
	}
	steps = append(steps, "4.capabilities")

	org, err = mgr.PutMemory(org.ID, "focus", "sample-A")
	if err != nil {
		t.Fatal(err)
	}
	if org.Memory.Working["focus"] != "sample-A" {
		t.Fatal("3.memory failed")
	}
	steps = append(steps, "3.memory")

	org, err = mgr.Start(org.ID)
	if err != nil || org.Status != organism.StatusRunning {
		t.Fatalf("5.start: %v %+v", err, org)
	}
	steps = append(steps, "5.start")

	// 6 Observe
	got, err := mgr.Get(org.ID)
	if err != nil || got.Status != organism.StatusRunning {
		t.Fatal("6.observe")
	}
	steps = append(steps, "6.observe")

	// 7 Send command (execution)
	eng := execution.BuiltinEngine{}
	res, err := eng.Execute(context.Background(), execution.Request{
		OrganismID: org.ID, Entry: "ping",
	})
	if err != nil || string(res.Output) != "pong" {
		t.Fatalf("7.command: %v %s", err, res.Output)
	}
	steps = append(steps, "7.command")

	// 8 Receive Pulse (bridge from events)
	if node.Pulses().Count() < 1 {
		t.Fatal("8.pulse: expected pulses from create/start")
	}
	foundStart := false
	for _, p := range node.Pulses().Recent(0) {
		if p.Type == "organism.started" {
			foundStart = true
		}
	}
	if !foundStart {
		t.Fatal("8.pulse: missing organism.started")
	}
	steps = append(steps, "8.pulse")

	// 9 Persist + restore
	node2, err := core.NewNode(core.Config{DataDir: filepath.Join(dir, "node"), Name: "e2e-node"})
	if err != nil {
		t.Fatal(err)
	}
	restored, err := node2.Organisms().Get(org.ID)
	if err != nil || restored.Memory.Working["focus"] != "sample-A" {
		t.Fatalf("9.persist: %v %+v", err, restored)
	}
	if restored.RootCID != org.RootCID {
		t.Fatal("9.persist RootCID mismatch")
	}
	steps = append(steps, "9.persist")

	// 10–12 Move / failure / recovery via local fabric
	fab := network.NewFabric()
	ta := network.NewLocalTransport("node-A", fab)
	tb := network.NewLocalTransport("node-B", fab)
	_ = ta.Listen(context.Background())
	_ = tb.Listen(context.Background())
	sa := replication.NewService("node-A", ta)
	sb := replication.NewService("node-B", tb)
	tb.OnMessage(sb.HandleMessage)
	ta.OnMessage(sa.HandleMessage)

	moving := restored.Snapshot()
	moving.Placement.Primary = "node-A"
	moving.Status = organism.StatusRunning
	_, err = sa.Replicate(moving, []string{"node-B"})
	if err != nil {
		t.Fatal(err)
	}
	if sb.GetReplica(moving.ID) == nil {
		t.Fatal("10.move/replicate: replica missing on B")
	}
	steps = append(steps, "10.replicate")

	online := map[string]bool{"node-A": false, "node-B": true}
	sb.SetOnlineFunc(func(id string) bool { return online[id] })
	_ = ta.Close()
	steps = append(steps, "11.failure")

	rr, err := sb.RecoverIfPrimaryDown(moving.ID)
	if err != nil || !rr.Success || rr.NewPrimary != "node-B" {
		t.Fatalf("12.recover: %+v %v", rr, err)
	}
	rec := sb.GetReplica(moving.ID)
	if rec.Organism.RootCID != org.RootCID || rec.Primary != "node-B" {
		t.Fatalf("12.recover state: %+v", rec)
	}
	steps = append(steps, "12.recover")

	// Inference smoke
	inf, err := inference.EchoProvider{}.Infer(context.Background(), inference.Request{Prompt: "ok"})
	if err != nil || inf.Text == "" {
		t.Fatal("inference")
	}

	t.Logf("E2E checklist OK: %v", steps)
	if len(steps) < 12 {
		t.Fatalf("expected >=12 steps, got %d %v", len(steps), steps)
	}
}

func TestE2E_AIP_HTTP(t *testing.T) {
	dir := t.TempDir()
	node, err := core.NewNode(core.Config{DataDir: dir, Name: "aip-e2e"})
	if err != nil {
		t.Fatal(err)
	}
	_ = node.Start(context.Background())
	defer node.Stop()

	srv := &aip.Server{Node: node}
	h := srv.Handler()

	post := func(action string, params map[string]any) aip.CommandResult {
		body, _ := json.Marshal(aip.Command{AIP: aip.Version, Action: action, Params: params})
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/aip/v1/commands", bytes.NewReader(body))
		h.ServeHTTP(rr, req)
		var res aip.CommandResult
		_ = json.Unmarshal(rr.Body.Bytes(), &res)
		if rr.Code != 200 && !res.OK {
			t.Fatalf("%s: %d %s", action, rr.Code, rr.Body.String())
		}
		return res
	}

	cr := post("create", map[string]any{
		"name": "http-e2e", "capabilities": []any{"memory.read", "memory.write", "inference"},
	})
	if cr.Organism == nil {
		t.Fatal("create")
	}
	id := cr.Organism.ID
	post("start", map[string]any{"id": id})
	post("memory.set", map[string]any{"id": id, "key": "k", "value": "v"})
	post("execute", map[string]any{"id": id, "entry": "ping"})
	post("infer", map[string]any{"prompt": "hello"})
	post("policy.check", map[string]any{"id": id, "action": "memory.read"})

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/aip/v1/info", nil))
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/aip/v1/pulses", nil))
	if rr.Code != 200 {
		t.Fatal("pulses")
	}

	// allow event bridge to settle
	time.Sleep(5 * time.Millisecond)
	if node.Pulses().Count() < 1 {
		t.Fatal("expected pulses")
	}
}
