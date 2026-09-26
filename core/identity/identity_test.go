package identity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if id.ID == "" {
		t.Fatal("expected non-empty NodeID")
	}
	if len(id.PublicKey) == 0 || len(id.PrivateKey) == 0 {
		t.Fatal("expected keys")
	}
	if id.ID[:5] != "node:" {
		t.Fatalf("NodeID should start with node:, got %s", id.ID)
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "identity.json")

	id, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := id.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ID != id.ID {
		t.Fatalf("ID mismatch: %s vs %s", loaded.ID, id.ID)
	}
}

func TestLoadOrCreate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "node", "identity.json")

	id1, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate (create): %v", err)
	}

	id2, err := LoadOrCreate(path)
	if err != nil {
		t.Fatalf("LoadOrCreate (load): %v", err)
	}
	if id1.ID != id2.ID {
		t.Fatalf("expected same ID, got %s and %s", id1.ID, id2.ID)
	}
}

func TestSignVerify(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	msg := []byte("prismatec-core-test")
	sig := id.Sign(msg)
	if !Verify(id.PublicKey, msg, sig) {
		t.Fatal("signature should verify")
	}
	if Verify(id.PublicKey, []byte("tampered"), sig) {
		t.Fatal("tampered message should not verify")
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !os.IsNotExist(err) && err.Error() == "" {
		// wrapped error is fine
	}
}
