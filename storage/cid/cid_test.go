package cid

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestNewDeterministic(t *testing.T) {
	a := New([]byte("hello"))
	b := New([]byte("hello"))
	c := New([]byte("world"))
	if a != b {
		t.Fatal("same content must same cid")
	}
	if a == c {
		t.Fatal("different content must different cid")
	}
	if !Valid(a) {
		t.Fatalf("invalid format: %s", a)
	}
}

func TestLocalStorePutGet(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "blocks")
	s, err := NewLocalStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("organism memory payload")
	id, err := s.Put(content)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Has(id) {
		t.Fatal("Has should be true")
	}
	got, err := s.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("got %q", got)
	}
	// idempotent
	id2, err := s.Put(content)
	if err != nil || id2 != id {
		t.Fatalf("idempotent put: %v %s", err, id2)
	}
}

func TestGetMissing(t *testing.T) {
	s, _ := NewLocalStore(t.TempDir())
	_, err := s.Get(New([]byte("nope")))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewIPFSAndPutIPFS(t *testing.T) {
	content := []byte("ipfs-style content")
	id, err := NewIPFS(content)
	if err != nil {
		t.Fatal(err)
	}
	if !Valid(id) {
		t.Fatalf("Valid failed for %s", id)
	}
	if len(id) > 4 && id[:5] == Prefix {
		t.Fatalf("IPFS cid should not use cid1 prefix: %s", id)
	}
	s, err := NewLocalStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gotID, err := s.PutIPFS(content)
	if err != nil {
		t.Fatal(err)
	}
	if gotID != id {
		t.Fatalf("%s vs %s", gotID, id)
	}
	data, err := s.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Fatalf("data mismatch")
	}
}
