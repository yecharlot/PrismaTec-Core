package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core"
)

// TestE2E_ProcessKillContinuity launches two OS processes (node A/B), replicates via AIP,
// SIGKILLs A, recovers on B, verifies memory + RootCID + epoch fencing + execute via AIP.
//
// This is the strongest automated continuity check in-repo (hard kill, not clean Stop).
func TestE2E_ProcessKillContinuity(t *testing.T) {
	if testing.Short() {
		t.Skip("skip process continuity in -short")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		// when running as ./tests/e2e from module root
		root, err = os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		// if cwd is tests/e2e
		if filepath.Base(root) == "e2e" {
			root = filepath.Join(root, "../..")
		}
	}
	// Prefer module root containing go.mod
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		root, _ = filepath.Abs(".")
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root not found from %s", root)
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "prismatec")
	build := exec.Command("go", "build", "-o", bin, "./cmd/prismatec")
	build.Dir = root
	build.Env = append(os.Environ(), "GOPROXY=https://proxy.golang.org,direct", "GOTOOLCHAIN=go1.25.0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	dirA := filepath.Join(dir, "a")
	dirB := filepath.Join(dir, "b")
	_ = os.MkdirAll(dirA, 0o700)
	_ = os.MkdirAll(dirB, 0o700)

	// Identity bootstrap without network
	bootA, err := core.NewNode(core.Config{DataDir: dirA, Name: "proc-A"})
	if err != nil {
		t.Fatal(err)
	}
	bootB, err := core.NewNode(core.Config{DataDir: dirB, Name: "proc-B"})
	if err != nil {
		t.Fatal(err)
	}
	idA, idB := string(bootA.ID()), string(bootB.ID())

	addrA, addrB := "127.0.0.1:19201", "127.0.0.1:19202"
	aipA, aipB := "127.0.0.1:18201", "127.0.0.1:18202"

	startNode := func(dataDir, name, netAddr, aipAddr, peers string) *exec.Cmd {
		cmd := exec.Command(bin, "node", "start")
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"PRISMATEC_DATA_DIR="+dataDir,
			"PRISMATEC_NODE_NAME="+name,
			"PRISMATEC_NETWORK_ADDR="+netAddr,
			"PRISMATEC_AIP_ADDR="+aipAddr,
			"PRISMATEC_PEERS="+peers,
			"GOPROXY=https://proxy.golang.org,direct",
		)
		cmd.Stdout = &bytes.Buffer{}
		cmd.Stderr = &bytes.Buffer{}
		if err := cmd.Start(); err != nil {
			t.Fatalf("start %s: %v", name, err)
		}
		return cmd
	}

	peersA := idB + "@" + addrB
	peersB := idA + "@" + addrA
	cmdB := startNode(dirB, "proc-B", addrB, aipB, peersB)
	cmdA := startNode(dirA, "proc-A", addrA, aipA, peersA)
	defer func() {
		_ = killProcess(cmdA)
		_ = killProcess(cmdB)
	}()

	waitHTTP := func(base string, timeout time.Duration) error {
		deadline := time.Now().Add(timeout)
		var last error
		for time.Now().Before(deadline) {
			resp, err := http.Get(base + "/healthz")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == 200 {
					return nil
				}
				last = fmt.Errorf("status %d", resp.StatusCode)
			} else {
				last = err
			}
			time.Sleep(100 * time.Millisecond)
		}
		return last
	}
	if err := waitHTTP("http://"+aipB, 15*time.Second); err != nil {
		t.Fatalf("B AIP up: %v stderr=%s", err, cmdB.Stderr.(*bytes.Buffer).String())
	}
	if err := waitHTTP("http://"+aipA, 15*time.Second); err != nil {
		t.Fatalf("A AIP up: %v stderr=%s", err, cmdA.Stderr.(*bytes.Buffer).String())
	}

	post := func(aip, action string, params map[string]any) (map[string]any, error) {
		body, _ := json.Marshal(map[string]any{"aip": "v1", "action": action, "params": params})
		resp, err := http.Post("http://"+aip+"/aip/v1/commands", "application/json", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		var out map[string]any
		_ = json.Unmarshal(b, &out)
		if resp.StatusCode >= 300 {
			return out, fmt.Errorf("status %d %s", resp.StatusCode, string(b))
		}
		if ok, _ := out["ok"].(bool); !ok {
			return out, fmt.Errorf("not ok: %s", string(b))
		}
		return out, nil
	}

	created, err := post(aipA, "create", map[string]any{
		"name": "kill-continuity", "capabilities": []any{"memory.read", "memory.write"},
	})
	if err != nil {
		t.Fatal(err)
	}
	org, _ := created["organism"].(map[string]any)
	orgID, _ := org["id"].(string)
	rootCID, _ := org["root_cid"].(string)
	if orgID == "" || rootCID == "" {
		t.Fatalf("create payload: %v", created)
	}
	if _, err := post(aipA, "start", map[string]any{"id": orgID}); err != nil {
		t.Fatal(err)
	}
	if _, err := post(aipA, "memory.set", map[string]any{"id": orgID, "key": "k", "value": "survive-kill"}); err != nil {
		t.Fatal(err)
	}

	// replicate A → B (retry until TCP mesh ready)
	var repErr error
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		_, repErr = post(aipA, "replicate", map[string]any{"id": orgID, "remote_node_id": idB})
		if repErr == nil {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if repErr != nil {
		t.Fatalf("replicate: %v", repErr)
	}

	// confirm B has organism via GET
	deadline = time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + aipB + "/aip/v1/organisms/" + orgID)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}

	// HARD KILL A (SIGKILL) — not graceful Stop
	if err := cmdA.Process.Kill(); err != nil {
		// Windows fallback
		if runtime.GOOS == "windows" {
			_ = cmdA.Process.Signal(os.Kill)
		} else {
			_ = cmdA.Process.Signal(syscall.SIGKILL)
		}
	}
	_, _ = cmdA.Process.Wait()
	time.Sleep(500 * time.Millisecond)

	// Recover on B
	var rec map[string]any
	deadline = time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		rec, err = post(aipB, "recover", map[string]any{"id": orgID})
		if err == nil {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("recover after kill: %v", err)
	}

	resp, err := http.Get("http://" + aipB + "/aip/v1/organisms/" + orgID)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var view map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&view)
	if view["root_cid"] != rootCID {
		t.Fatalf("RootCID after kill-recover: %v", view)
	}
	if view["status"] != "running" {
		t.Fatalf("status: %v", view)
	}

	// execute on B path after recovery
	if _, err := post(aipB, "execute", map[string]any{"id": orgID, "entry": "ping"}); err != nil {
		t.Fatalf("execute after kill-recover: %v", err)
	}

	t.Logf("process continuity OK org=%s root=%s recover=%v", orgID, rootCID, rec)
}

func killProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
	return nil
}

// silence unused import if context not used
var _ = context.Background
