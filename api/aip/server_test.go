package aip

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
)

func testNode(t *testing.T) *core.Node {
	t.Helper()
	n, err := core.NewNode(core.Config{
		DataDir: filepath.Join(t.TempDir(), "node"),
		Name:    "aip-test-node",
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestInfoAndOrganisms(t *testing.T) {
	node := testNode(t)
	_, err := node.Organisms().Create(organism.CreateOptions{
		Name:         "aip-agent",
		Capabilities: []organism.Capability{"memory.read"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{Node: node}
	h := s.Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/aip/v1/info", nil))
	if rr.Code != 200 {
		t.Fatalf("info status %d", rr.Code)
	}
	var info NodeInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.AIP != Version || info.Organisms != 1 {
		t.Fatalf("info = %+v", info)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/aip/v1/organisms", nil))
	if rr.Code != 200 {
		t.Fatalf("organisms status %d body %s", rr.Code, rr.Body.String())
	}
}

func TestCommandsCreateStart(t *testing.T) {
	node := testNode(t)
	s := &Server{Node: node}
	h := s.Handler()

	body, _ := json.Marshal(Command{
		AIP: Version, Action: "create",
		Params: map[string]any{"name": "cmd-agent", "capabilities": []any{"memory.read", "inference"}},
	})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/aip/v1/commands", bytes.NewReader(body)))
	if rr.Code != 200 {
		t.Fatalf("create %d %s", rr.Code, rr.Body.String())
	}
	var res CommandResult
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	if !res.OK || res.Organism == nil {
		t.Fatalf("res %+v", res)
	}

	body, _ = json.Marshal(Command{
		AIP: Version, Action: "start",
		Params: map[string]any{"id": res.Organism.ID},
	})
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/aip/v1/commands", bytes.NewReader(body)))
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	if !res.OK || res.Organism.Status != "running" {
		t.Fatalf("start %+v", res)
	}
}

func TestPulsesRecentAfterCreate(t *testing.T) {
	node := testNode(t)
	_, _ = node.Organisms().Create(organism.CreateOptions{Name: "pulse-via-aip"})
	s := &Server{Node: node}
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/aip/v1/pulses", nil))
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	var out struct {
		Pulses []PulseMessage `json:"pulses"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if len(out.Pulses) < 1 {
		t.Fatal("expected at least one pulse from create")
	}
}
