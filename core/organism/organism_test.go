package organism

import (
	"testing"

	"github.com/yecharlot/PrismaTec-Core/core/events"
	"github.com/yecharlot/PrismaTec-Core/core/registry"
)

func TestNewAndRootCID(t *testing.T) {
	o1, err := New(CreateOptions{Name: "research-agent-01", Capabilities: []Capability{"memory.read", "inference"}})
	if err != nil {
		t.Fatal(err)
	}
	o2, err := New(CreateOptions{Name: "research-agent-01", Capabilities: []Capability{"memory.read", "inference"}})
	if err != nil {
		t.Fatal(err)
	}
	if o1.RootCID != o2.RootCID {
		t.Fatalf("same manifest should yield same RootCID: %s vs %s", o1.RootCID, o2.RootCID)
	}
	if o1.RootCID[:8] != "rootcid:" {
		t.Fatalf("RootCID prefix: %s", o1.RootCID)
	}
	if !o1.HasCapability("memory.read") {
		t.Fatal("expected capability")
	}
	if !o1.Allowed("memory.read") {
		t.Fatal("expected allowed by default policy")
	}
}

func TestRootCIDDiffersByCapability(t *testing.T) {
	a, _ := New(CreateOptions{Name: "x", Capabilities: []Capability{"a"}})
	b, _ := New(CreateOptions{Name: "x", Capabilities: []Capability{"b"}})
	if a.RootCID == b.RootCID {
		t.Fatal("different capabilities should change RootCID")
	}
}

func TestManagerLifecycle(t *testing.T) {
	dir := t.TempDir()
	bus := events.NewBus()
	reg := registry.New()
	var created, started, stopped int
	bus.Subscribe("organism.created", func(e events.Event) { created++ })
	bus.Subscribe("organism.started", func(e events.Event) { started++ })
	bus.Subscribe("organism.stopped", func(e events.Event) { stopped++ })

	m, err := NewManager(dir, "node:test", "test-node", bus, reg)
	if err != nil {
		t.Fatal(err)
	}

	org, err := m.Create(CreateOptions{
		Name:         "research-agent-01",
		Capabilities: []Capability{"memory.read", "inference"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if org.Status != StatusReady {
		t.Fatalf("status = %s", org.Status)
	}
	if created != 1 {
		t.Fatalf("created events = %d", created)
	}

	org, err = m.Start(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if org.Status != StatusRunning {
		t.Fatalf("status = %s", org.Status)
	}
	if started != 1 {
		t.Fatalf("started = %d", started)
	}

	org, err = m.Stop(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if org.Status != StatusStopped {
		t.Fatalf("status = %s", org.Status)
	}
	if stopped != 1 {
		t.Fatalf("stopped = %d", stopped)
	}

	// persistence + reload
	m2, err := NewManager(dir, "node:test", "test-node", events.NewBus(), registry.New())
	if err != nil {
		t.Fatal(err)
	}
	got, err := m2.Get(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "research-agent-01" || got.RootCID != org.RootCID {
		t.Fatalf("reload mismatch: %+v", got)
	}
	if len(m2.List()) != 1 {
		t.Fatalf("list = %d", len(m2.List()))
	}

	// get by name
	byName, err := m2.Get("research-agent-01")
	if err != nil || byName.ID != org.ID {
		t.Fatalf("get by name: %v %+v", err, byName)
	}
}

func TestManagerDuplicateName(t *testing.T) {
	dir := t.TempDir()
	m, err := NewManager(dir, "node:t", "n", events.NewBus(), registry.New())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Create(CreateOptions{Name: "a"}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Create(CreateOptions{Name: "a"}); err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func TestPutMemory(t *testing.T) {
	dir := t.TempDir()
	m, _ := NewManager(dir, "node:t", "n", events.NewBus(), registry.New())
	org, _ := m.Create(CreateOptions{Name: "mem-agent"})
	org, err := m.PutMemory(org.ID, "sample", "analyzed")
	if err != nil {
		t.Fatal(err)
	}
	if org.Memory.Working["sample"] != "analyzed" {
		t.Fatalf("memory = %+v", org.Memory.Working)
	}
	m2, err := NewManager(dir, "node:t", "n", events.NewBus(), registry.New())
	if err != nil {
		t.Fatal(err)
	}
	got, err := m2.Get(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Memory.Working["sample"] != "analyzed" {
		t.Fatalf("persisted memory = %+v", got.Memory.Working)
	}
}

func TestEpisodicAndCID(t *testing.T) {
	dir := t.TempDir()
	m, err := NewManager(dir, "node:t", "n", events.NewBus(), registry.New())
	if err != nil {
		t.Fatal(err)
	}
	org, err := m.Create(CreateOptions{Name: "lab-agent"})
	if err != nil {
		t.Fatal(err)
	}
	blob := []byte("observation: sample A positive")
	org, err = m.AppendEpisode(org.ID, "observation", map[string]any{"source": "lab"}, blob)
	if err != nil {
		t.Fatal(err)
	}
	if len(org.Memory.Episodic) != 1 {
		t.Fatalf("episodes = %d", len(org.Memory.Episodic))
	}
	ep := org.Memory.Episodic[0]
	if ep.ContentCID == "" {
		t.Fatal("expected ContentCID")
	}
	data, err := m.GetBlock(ep.ContentCID)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(blob) {
		t.Fatalf("block = %q", data)
	}
	if _, err = m.PutSemantic(org.ID, "protocol", "PCR-v2"); err != nil {
		t.Fatal(err)
	}
	m2, err := NewManager(dir, "node:t", "n", events.NewBus(), registry.New())
	if err != nil {
		t.Fatal(err)
	}
	got, err := m2.Get(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Memory.Episodic) != 1 || got.Memory.Semantic["protocol"] != "PCR-v2" {
		t.Fatalf("restore memory: %+v", got.Memory)
	}
	data2, err := m2.GetBlock(got.Memory.Episodic[0].ContentCID)
	if err != nil || string(data2) != string(blob) {
		t.Fatalf("restore block: %v %q", err, data2)
	}
}
