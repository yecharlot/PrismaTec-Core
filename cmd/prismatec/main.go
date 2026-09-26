package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/yecharlot/PrismaTec-Core/api/aip"
	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/network"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "node":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: prismatec node start|info")
			os.Exit(1)
		}
		switch os.Args[2] {
		case "start":
			runNodeStart()
		case "info":
			runNodeInfo()
		default:
			fmt.Fprintf(os.Stderr, "unknown node command: %s\n", os.Args[2])
			os.Exit(1)
		}
	case "organism":
		runOrganism(os.Args[2:])
	case "pulse":
		runPulse(os.Args[2:])
	case "demo":
		runDemo(os.Args[2:])
	case "version":
		fmt.Println("prismatec-core 0.2.0-dev (phase-14-e2e)")
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`PrismaTec Core — infrastructure runtime for persistent autonomous digital entities

Usage:
  prismatec node start              Start a local Core node
  prismatec node info               Show node identity
  prismatec organism create <name>  Create an organism
  prismatec organism list           List organisms
  prismatec organism inspect <id>   Inspect organism (id or name)
  prismatec organism start <id>     Start organism
  prismatec organism stop <id>      Stop organism
  prismatec organism memory set <id> <key> <value>
  prismatec organism memory semantic <id> <key> <value>
  prismatec organism memory episode <id> <type> [content...]
  prismatec pulse list [n]           Show recent pulses (default 20)
  prismatec demo e2e               Run end-to-end checklist (Phase 14)
  prismatec version
  prismatec help

Flags for create:
  --cap <name>     Add capability (repeatable). Default: memory.read,inference
  --deny <name>    Deny capability in policy

Env:
  PRISMATEC_DATA_DIR    Data directory (identity + organisms)
  PRISMATEC_NODE_NAME   Human-readable node name`)
}

func dataDir() string {
	if d := os.Getenv("PRISMATEC_DATA_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".prismatec", "node")
	}
	return filepath.Join(home, ".prismatec", "node")
}

func openNode() *core.Node {
	cfg := core.DefaultConfig()
	cfg.DataDir = dataDir()
	if name := os.Getenv("PRISMATEC_NODE_NAME"); name != "" {
		cfg.Name = name
	}
	if a := os.Getenv("PRISMATEC_NETWORK_ADDR"); a != "" {
		cfg.NetworkAddr = a
	}
	if peers := os.Getenv("PRISMATEC_PEERS"); peers != "" {
		cfg.Peers = network.ParsePeersEnv(peers)
	}
	if tk := os.Getenv("PRISMATEC_TRANSPORT"); tk != "" {
		cfg.TransportKind = tk
	}
	node, err := core.NewNode(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return node
}

func aipAddr() string {
	if a := os.Getenv("PRISMATEC_AIP_ADDR"); a != "" {
		return a
	}
	return ":8080"
}

func runNodeStart() {
	node := openNode()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := node.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error starting node: %v\n", err)
		os.Exit(1)
	}

	addr := aipAddr()
	srv := &aip.Server{Node: node, Addr: addr}
	httpServer := &http.Server{Addr: addr, Handler: srv.Handler()}
	go func() {
		fmt.Printf("  AIP v1 listening on http://localhost%s\n", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "AIP server error: %v\n", err)
		}
	}()

	info := node.Info()
	fmt.Println("══════════════════════════════════════════")
	fmt.Println("         PRISMATEC CORE  —  NODE")
	fmt.Println("══════════════════════════════════════════")
	fmt.Printf("  Name      %s\n", info["name"])
	fmt.Printf("  NodeID    %s\n", info["node_id"])
	fmt.Printf("  Status    %s\n", info["status"])
	fmt.Printf("  DataDir   %s\n", info["data_dir"])
	fmt.Printf("  Organisms %v\n", info["organisms"])
	fmt.Printf("  AIP       http://localhost%s/aip/v1/info\n", addr)
	fmt.Println("══════════════════════════════════════════")
	fmt.Println("  Node + AIP running. Press Ctrl+C to stop.")
	fmt.Println()

	<-ctx.Done()
	fmt.Println("\nShutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	_ = node.Stop()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Node stopped.")
}

func runNodeInfo() {
	node := openNode()
	info := node.Info()
	fmt.Printf("Name:      %s\n", info["name"])
	fmt.Printf("NodeID:    %s\n", info["node_id"])
	fmt.Printf("Status:    %s\n", info["status"])
	fmt.Printf("DataDir:   %s\n", info["data_dir"])
	fmt.Printf("Organisms: %v\n", info["organisms"])
}

func runOrganism(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: prismatec organism create|list|inspect|start|stop")
		os.Exit(1)
	}
	node := openNode()
	mgr := node.Organisms()

	switch args[0] {
	case "create":
		name, caps, deny := parseCreateArgs(args[1:])
		if name == "" {
			fmt.Fprintln(os.Stderr, "usage: prismatec organism create <name> [--cap x] [--deny y]")
			os.Exit(1)
		}
		policy := organism.Policy{Rules: map[string]bool{}}
		for _, c := range caps {
			policy.Rules[string(c)] = true
		}
		for _, d := range deny {
			policy.Rules[d] = false
		}
		org, err := mgr.Create(organism.CreateOptions{
			Name:         name,
			Capabilities: caps,
			Policy:       policy,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printOrganism(org)

	case "list":
		list := mgr.List()
		if len(list) == 0 {
			fmt.Println("(no organisms)")
			return
		}
		fmt.Printf("%-28s %-12s %-18s %s\n", "ID", "STATUS", "NAME", "ROOTCID")
		for _, o := range list {
			fmt.Printf("%-28s %-12s %-18s %s\n", o.ID, o.Status, o.Name, truncate(o.RootCID, 24))
		}

	case "inspect":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: prismatec organism inspect <id|name>")
			os.Exit(1)
		}
		org, err := mgr.Get(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printOrganism(org)

	case "start":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: prismatec organism start <id|name>")
			os.Exit(1)
		}
		org, err := mgr.Start(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printOrganism(org)

	case "stop":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: prismatec organism stop <id|name>")
			os.Exit(1)
		}
		org, err := mgr.Stop(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printOrganism(org)

	case "memory":
		runOrganismMemory(mgr, args[1:])

	default:
		fmt.Fprintf(os.Stderr, "unknown organism command: %s\n", args[0])
		os.Exit(1)
	}
}

func parseCreateArgs(args []string) (name string, caps []organism.Capability, deny []string) {
	caps = []organism.Capability{"memory.read", "inference"}
	customCaps := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--cap" && i+1 < len(args):
			if !customCaps {
				caps = nil
				customCaps = true
			}
			caps = append(caps, organism.Capability(args[i+1]))
			i++
		case a == "--deny" && i+1 < len(args):
			deny = append(deny, args[i+1])
			i++
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", a)
			os.Exit(1)
		default:
			if name == "" {
				name = a
			}
		}
	}
	return name, caps, deny
}

func printOrganism(o *organism.Organism) {
	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Println("│             PRISMATEC CORE                 │")
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Println("│  ORGANISM                                   │")
	fmt.Printf("│  ID: %-38s │\n", o.ID)
	fmt.Printf("│  RootCID: %-32s │\n", truncate(o.RootCID, 32))
	fmt.Printf("│  Node: %-36s │\n", truncate(o.Placement.Primary, 36))
	fmt.Println("│                                             │")
	fmt.Printf("│  STATUS       ● %-27s │\n", strings.ToUpper(string(o.Status)))
	fmt.Println("│                                             │")
	fmt.Printf("│  MEMORY       %d working · %d episodes · %d semantic │\n",
		len(o.Memory.Working), len(o.Memory.Episodic), len(o.Memory.Semantic))
	fmt.Printf("│  CAPABILITIES %d%-29s │\n", len(o.Capabilities), "")
	if o.CurrentAction != "" {
		fmt.Printf("│  CURRENT ACTION  %-26s │\n", truncate(o.CurrentAction, 26))
	}
	fmt.Println("│                                             │")
	fmt.Println("│  POLICY                                      │")
	for _, c := range o.Capabilities {
		mark := "✓"
		if !o.Allowed(c) {
			mark = "✕"
		}
		fmt.Printf("│  %s %-38s │\n", mark, c)
	}
	// also show explicit denies not in capabilities list
	for k, allowed := range o.Policy.Rules {
		if allowed {
			continue
		}
		found := false
		for _, c := range o.Capabilities {
			if string(c) == k {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("│  ✕ %-38s │\n", k)
		}
	}
	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Printf("Name: %s  Version: %s  Updated: %s\n", o.Name, o.Version, o.UpdatedAt.Format(time.RFC3339))
}


func printMultiNodeDemo() {
	fmt.Print(`PrismaTec Core — Demo 2 & 3 (multi-node TCP)

Terminal A (primary):
  export PRISMATEC_DATA_DIR=/tmp/ptc-a
  export PRISMATEC_NODE_NAME=node-A
  export PRISMATEC_NETWORK_ADDR=127.0.0.1:9001
  export PRISMATEC_AIP_ADDR=:8081
  # After B starts, set PEERS to B's NodeID@127.0.0.1:9002
  go run ./cmd/prismatec node start
  # note NodeID from banner / node info

Terminal B (replica):
  export PRISMATEC_DATA_DIR=/tmp/ptc-b
  export PRISMATEC_NODE_NAME=node-B
  export PRISMATEC_NETWORK_ADDR=127.0.0.1:9002
  export PRISMATEC_AIP_ADDR=:8082
  export PRISMATEC_PEERS="<NODE_A_ID>@127.0.0.1:9001"
  go run ./cmd/prismatec node start

Then:
  # on A — create & start organism, then replicate to B's NodeID
  PRISMATEC_DATA_DIR=/tmp/ptc-a go run ./cmd/prismatec organism create mover --cap memory.read
  PRISMATEC_DATA_DIR=/tmp/ptc-a go run ./cmd/prismatec organism start mover
  PRISMATEC_DATA_DIR=/tmp/ptc-a PRISMATEC_NETWORK_ADDR=127.0.0.1:9001 \
    PRISMATEC_PEERS="<NODE_B_ID>@127.0.0.1:9002" \
    go run ./cmd/prismatec organism replicate <orgId> <NODE_B_ID>

Demo 3 — stop A, recover on B:
  # Ctrl+C on A
  PRISMATEC_DATA_DIR=/tmp/ptc-b PRISMATEC_NETWORK_ADDR=127.0.0.1:9002 \
    go run ./cmd/prismatec organism recover <orgId>

Automated proof:
  go test ./tests/e2e/ -run MultiNode -count=1 -v
\n`)
}

func runDemo(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: prismatec demo e2e|multinode")
		os.Exit(1)
	}
	if args[0] == "multinode" {
		printMultiNodeDemo()
		return
	}
	if args[0] != "e2e" {
		fmt.Fprintln(os.Stderr, "usage: prismatec demo e2e|multinode")
		os.Exit(1)
	}
	// Delegate to documented test path — print instructions + quick local smoke via node APIs
	fmt.Println("PrismaTec Core — E2E Demo (Phase 14)")
	fmt.Println("Full automated checklist:")
	fmt.Println("  go test ./tests/e2e/ -count=1 -v")
	fmt.Println()
	fmt.Println("Manual Studio path:")
	fmt.Println("  1. go run ./cmd/prismatec node start")
	fmt.Println("  2. open http://127.0.0.1:8080/studio/")
	fmt.Println("  3. Create → Start → Memory → Pulse stream → Execute / Policy")
	fmt.Println()
	// Local smoke using same DataDir process
	node := openNode()
	mgr := node.Organisms()
	name := fmt.Sprintf("demo-%d", time.Now().Unix()%100000)
	org, err := mgr.Create(organism.CreateOptions{
		Name: name,
		Capabilities: []organism.Capability{"memory.read", "memory.write", "inference"},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[ok] create     id=%s rootcid=%s\n", org.ID, truncate(org.RootCID, 28))
	org, err = mgr.PutMemory(org.ID, "demo", "1")
	if err != nil {
		fmt.Fprintf(os.Stderr, "memory: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[ok] memory")
	org, err = mgr.Start(org.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[ok] start      status=%s\n", org.Status)
	fmt.Printf("[ok] pulses     count=%d (this process; full pulse log needs node start)\n", node.Pulses().Count())
	fmt.Println("[ok] identity   NodeID="+string(node.ID()))
	fmt.Println()
	fmt.Println("Checklist 10–12 (replicate/recover) covered by: go test ./tests/e2e/")
	fmt.Println("Studio: http://127.0.0.1:8080/studio/  after  prismatec node start")
}


func runPulse(args []string) {
	node := openNode()
	// generate activity so pulses exist if organisms change was prior process:
	// pulses are in-memory only in Phase 4 — list shows current process log.
	n := 20
	if len(args) >= 1 && args[0] == "list" {
		if len(args) >= 2 {
			fmt.Sscanf(args[1], "%d", &n)
		}
	} else if len(args) >= 1 && args[0] != "list" {
		fmt.Fprintln(os.Stderr, "usage: prismatec pulse list [n]")
		os.Exit(1)
	}
	// Touch bus by listing organisms via a no-op path: emit inspect pulse for demo
	for _, o := range node.Organisms().List() {
		node.Pulses().EmitType(o.ID, "organism.observed", string(node.ID()), map[string]any{
			"name": o.Name, "status": string(o.Status),
		})
	}
	recent := node.Pulses().Recent(n)
	if len(recent) == 0 {
		fmt.Println("(no pulses yet — start node or create/start organisms in this process)")
		return
	}
	fmt.Printf("Pulses (showing %d, seq=%d)\n", len(recent), node.Pulses().Seq())
	for _, p := range recent {
		fmt.Printf("  [%s] target=%s type=%s source=%s\n",
			p.Timestamp.Format("15:04:05"), truncate(p.Target, 24), p.Type, truncate(p.Source, 20))
	}
}


func runOrganismMemory(mgr *organism.Manager, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: prismatec organism memory set|semantic|episode ...")
		os.Exit(1)
	}
	switch args[0] {
	case "set":
		if len(args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: prismatec organism memory set <id> <key> <value>")
			os.Exit(1)
		}
		org, err := mgr.PutMemory(args[1], args[2], strings.Join(args[3:], " "))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printOrganism(org)
	case "semantic":
		if len(args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: prismatec organism memory semantic <id> <key> <value>")
			os.Exit(1)
		}
		org, err := mgr.PutSemantic(args[1], args[2], strings.Join(args[3:], " "))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printOrganism(org)
	case "episode":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: prismatec organism memory episode <id> <type> [content...]")
			os.Exit(1)
		}
		var content []byte
		if len(args) > 3 {
			content = []byte(strings.Join(args[3:], " "))
		}
		org, err := mgr.AppendEpisode(args[1], args[2], map[string]any{"via": "cli"}, content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printOrganism(org)
		if len(org.Memory.Episodic) > 0 {
			ep := org.Memory.Episodic[len(org.Memory.Episodic)-1]
			fmt.Printf("Episode: %s type=%s cid=%s\n", ep.ID, ep.Type, ep.ContentCID)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown memory command: %s\n", args[0])
		os.Exit(1)
	}
}


func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
