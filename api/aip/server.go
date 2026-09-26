package aip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/core/pulse"
)

// Server exposes AIP v1 over HTTP + SSE.
type Server struct {
	Node *core.Node
	Addr string // e.g. ":8080"
}

// Handler returns the HTTP mux for AIP v1.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/aip/v1/info", s.handleInfo)
	mux.HandleFunc("/aip/v1/organisms", s.handleOrganisms)
	mux.HandleFunc("/aip/v1/organisms/", s.handleOrganismByID)
	mux.HandleFunc("/aip/v1/pulse", s.handlePulseSSE)
	mux.HandleFunc("/aip/v1/pulses", s.handlePulsesRecent)
	mux.HandleFunc("/aip/v1/commands", s.handleCommands)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorBody{AIP: Version, Error: "method not allowed"})
		return
	}
	info := s.Node.Info()
	seq := uint64(0)
	pulses := 0
	if s.Node.Pulses() != nil {
		seq = s.Node.Pulses().Seq()
		pulses = s.Node.Pulses().Count()
	}
	orgCount := 0
	if n, ok := info["organisms"].(int); ok {
		orgCount = n
	}
	writeJSON(w, http.StatusOK, NodeInfo{
		AIP:       Version,
		Name:      fmt.Sprint(info["name"]),
		NodeID:    fmt.Sprint(info["node_id"]),
		Status:    fmt.Sprint(info["status"]),
		Organisms: orgCount,
		Pulses:    pulses,
		PulseSeq:  seq,
		DataDir:   fmt.Sprint(info["data_dir"]),
	})
}

func viewOf(o *organism.Organism) OrganismView {
	caps := make([]string, 0, len(o.Capabilities))
	for _, c := range o.Capabilities {
		caps = append(caps, string(c))
	}
	return OrganismView{
		ID:            o.ID,
		RootCID:       o.RootCID,
		Name:          o.Name,
		Status:        string(o.Status),
		NodeID:        o.Placement.Primary,
		Capabilities:  caps,
		Policy:        o.Policy.Rules,
		WorkingKeys:   len(o.Memory.Working),
		Episodes:      len(o.Memory.Episodic),
		SemanticKeys:  len(o.Memory.Semantic),
		CurrentAction: o.CurrentAction,
		UpdatedAt:     o.UpdatedAt,
	}
}

func (s *Server) handleOrganisms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorBody{AIP: Version, Error: "method not allowed"})
		return
	}
	list := s.Node.Organisms().List()
	out := make([]OrganismView, 0, len(list))
	for _, o := range list {
		out = append(out, viewOf(o))
	}
	writeJSON(w, http.StatusOK, map[string]any{"aip": Version, "organisms": out})
}

func (s *Server) handleOrganismByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorBody{AIP: Version, Error: "method not allowed"})
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/aip/v1/organisms/")
	id = strings.Trim(id, "/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorBody{AIP: Version, Error: "missing id"})
		return
	}
	o, err := s.Node.Organisms().Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, ErrorBody{AIP: Version, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, viewOf(o))
}

func (s *Server) handlePulsesRecent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorBody{AIP: Version, Error: "method not allowed"})
		return
	}
	n := 50
	recent := s.Node.Pulses().Recent(n)
	msgs := make([]PulseMessage, 0, len(recent))
	for _, p := range recent {
		msgs = append(msgs, PulseMessage{
			AIP: Version, Target: p.Target, Type: p.Type,
			Data: p.Data, Meta: p.Meta, Source: p.Source, Timestamp: p.Timestamp,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"aip": Version, "pulses": msgs})
}

// handlePulseSSE streams pulses as Server-Sent Events.
func (s *Server) handlePulseSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorBody{AIP: Version, Error: "method not allowed"})
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, ErrorBody{AIP: Version, Error: "streaming unsupported"})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan pulse.Pulse, 32)
	s.Node.Pulses().Subscribe(func(p pulse.Pulse) {
		select {
		case ch <- p:
		default:
			// drop if slow client
		}
	})

	// hello event
	fmt.Fprintf(w, "event: aip.hello\ndata: {\"aip\":\"%s\",\"node_id\":%q}\n\n", Version, s.Node.ID())
	flusher.Flush()

	ctx := r.Context()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case p := <-ch:
			msg := PulseMessage{
				AIP: Version, Target: p.Target, Type: p.Type,
				Data: p.Data, Meta: p.Meta, Source: p.Source, Timestamp: p.Timestamp,
			}
			b, _ := json.Marshal(msg)
			fmt.Fprintf(w, "event: pulse\ndata: %s\n\n", b)
			flusher.Flush()
		}
	}
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorBody{AIP: Version, Error: "method not allowed"})
		return
	}
	var cmd Command
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		writeJSON(w, http.StatusBadRequest, CommandResult{AIP: Version, OK: false, Error: "invalid json"})
		return
	}
	if cmd.AIP != "" && cmd.AIP != Version {
		writeJSON(w, http.StatusBadRequest, CommandResult{AIP: Version, OK: false, Error: "unsupported aip version"})
		return
	}
	res := s.execCommand(cmd)
	status := http.StatusOK
	if !res.OK {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, res)
}

func (s *Server) execCommand(cmd Command) CommandResult {
	mgr := s.Node.Organisms()
	params := cmd.Params
	if params == nil {
		params = map[string]any{}
	}
	str := func(k string) string {
		if v, ok := params[k].(string); ok {
			return v
		}
		return ""
	}

	switch cmd.Action {
	case "create":
		name := str("name")
		if name == "" {
			return CommandResult{AIP: Version, OK: false, Error: "params.name required"}
		}
		caps := []organism.Capability{"memory.read"}
		if raw, ok := params["capabilities"].([]any); ok {
			caps = nil
			for _, x := range raw {
				if s, ok := x.(string); ok {
					caps = append(caps, organism.Capability(s))
				}
			}
		}
		o, err := mgr.Create(organism.CreateOptions{Name: name, Capabilities: caps})
		if err != nil {
			return CommandResult{AIP: Version, OK: false, Error: err.Error()}
		}
		v := viewOf(o)
		return CommandResult{AIP: Version, OK: true, Organism: &v}

	case "start":
		id := str("id")
		if id == "" {
			id = str("name")
		}
		o, err := mgr.Start(id)
		if err != nil {
			return CommandResult{AIP: Version, OK: false, Error: err.Error()}
		}
		v := viewOf(o)
		return CommandResult{AIP: Version, OK: true, Organism: &v}

	case "stop":
		id := str("id")
		if id == "" {
			id = str("name")
		}
		o, err := mgr.Stop(id)
		if err != nil {
			return CommandResult{AIP: Version, OK: false, Error: err.Error()}
		}
		v := viewOf(o)
		return CommandResult{AIP: Version, OK: true, Organism: &v}

	case "memory.set":
		id := str("id")
		key := str("key")
		val := str("value")
		if id == "" || key == "" {
			return CommandResult{AIP: Version, OK: false, Error: "params.id and params.key required"}
		}
		o, err := mgr.PutMemory(id, key, val)
		if err != nil {
			return CommandResult{AIP: Version, OK: false, Error: err.Error()}
		}
		v := viewOf(o)
		return CommandResult{AIP: Version, OK: true, Organism: &v}

	default:
		return CommandResult{AIP: Version, OK: false, Error: "unknown action: " + cmd.Action}
	}
}
