package wasm

import (
	"context"
	"testing"

	"github.com/yecharlot/PrismaTec-Core/runtime/execution"
)

func TestInstantiateEmptyModule(t *testing.T) {
	// minimal WASM module (magic + version only)
	code := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	e := NewEngine()
	res, err := e.Execute(context.Background(), execution.Request{Input: code, Entry: "alset_main"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Engine != "wasm" {
		t.Fatalf("%+v", res)
	}
}
