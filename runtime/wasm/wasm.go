package wasm

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/yecharlot/PrismaTec-Core/runtime/execution"
)

// Engine runs WASM modules via wazero (sandboxed runtime).
type Engine struct {
	DefaultTimeout time.Duration
}

func NewEngine() *Engine {
	return &Engine{DefaultTimeout: 5 * time.Second}
}

func (e *Engine) Name() string { return "wasm" }

func (e *Engine) Execute(ctx context.Context, req execution.Request) (execution.Result, error) {
	start := time.Now()
	timeout := e.DefaultTimeout
	if req.Limits.Timeout > 0 {
		timeout = req.Limits.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var code []byte
	var err error
	if len(req.Input) > 4 && string(req.Input[:4]) == "\x00asm" {
		code = req.Input
	} else if req.Module != "" {
		code, err = os.ReadFile(req.Module)
		if err != nil {
			return execution.Result{}, fmt.Errorf("read wasm: %w", err)
		}
	} else {
		return execution.Result{}, fmt.Errorf("wasm requires module path or wasm bytes in input")
	}

	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	mod, err := rt.InstantiateWithConfig(ctx, code, wazero.NewModuleConfig().WithStartFunctions())
	if err != nil {
		return execution.Result{}, fmt.Errorf("instantiate wasm: %w", err)
	}
	entry := req.Entry
	if entry == "" {
		entry = "alset_main"
	}
	fn := mod.ExportedFunction(entry)
	if fn == nil {
		// module loaded OK but no export — still counts as successful load for empty modules
		return execution.Result{
			Output:   []byte("instantiated"),
			Engine:   "wasm",
			Duration: time.Since(start),
			Logs:     "no export " + entry,
		}, nil
	}
	vals, err := fn.Call(ctx)
	if err != nil {
		return execution.Result{}, fmt.Errorf("call wasm: %w", err)
	}
	out := []byte(fmt.Sprintf("%v", vals))
	return execution.Result{Output: out, Engine: "wasm", Duration: time.Since(start)}, nil
}
