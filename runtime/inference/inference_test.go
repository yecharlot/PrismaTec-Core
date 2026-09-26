package inference

import (
	"context"
	"testing"
)

func TestEchoProvider(t *testing.T) {
	r := NewRegistry()
	r.Register(EchoProvider{})
	p, err := r.Get("")
	if err != nil {
		t.Fatal(err)
	}
	res, err := p.Infer(context.Background(), Request{Prompt: "hello"})
	if err != nil || res.Text != "[echo] hello" {
		t.Fatalf("%+v %v", res, err)
	}
}
