package execution

import (
	"context"
	"testing"
)

func TestBuiltinEcho(t *testing.T) {
	reg := NewRegistry()
	reg.Register(BuiltinEngine{})
	e, err := reg.Get("builtin")
	if err != nil {
		t.Fatal(err)
	}
	res, err := e.Execute(context.Background(), Request{Entry: "echo", Input: []byte("hi")})
	if err != nil || string(res.Output) != "hi" {
		t.Fatalf("%v %v", res, err)
	}
	res, err = e.Execute(context.Background(), Request{Entry: "ping"})
	if err != nil || string(res.Output) != "pong" {
		t.Fatalf("%v %v", res, err)
	}
}
