// Package identity provides NodeID (persistent node identity) based on Ed25519.
// RootCID and PeerID are separate concerns and must never be mixed with NodeID.
package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// NodeID is the persistent cryptographic identity of a PrismaTec Core node.
// It is distinct from RootCID (organism content identity) and PeerID (libp2p transport identity).
type NodeID string

// Identity holds an Ed25519 keypair and its derived NodeID.
type Identity struct {
	ID         NodeID
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

type document struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

// New generates a fresh Ed25519 identity and derives its NodeID.
func New() (*Identity, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate identity: %w", err)
	}
	return fromKeys(pub, priv), nil
}

// Load reads a previously saved identity from path.
func Load(path string) (*Identity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read identity: %w", err)
	}
	var doc document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse identity: %w", err)
	}
	pub, err := base64.StdEncoding.DecodeString(doc.Public)
	if err != nil {
		return nil, fmt.Errorf("decode public key: %w", err)
	}
	priv, err := base64.StdEncoding.DecodeString(doc.Private)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	if len(pub) != ed25519.PublicKeySize || len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size")
	}
	return fromKeys(ed25519.PublicKey(pub), ed25519.PrivateKey(priv)), nil
}

// LoadOrCreate loads identity from path, or creates and saves a new one if missing.
func LoadOrCreate(path string) (*Identity, error) {
	id, err := Load(path)
	if err == nil {
		return id, nil
	}
	// Create only when the file does not exist.
	if _, statErr := os.Stat(path); statErr == nil {
		return nil, err // file exists but Load failed
	}
	id, err = New()
	if err != nil {
		return nil, err
	}
	if err := id.Save(path); err != nil {
		return nil, err
	}
	return id, nil
}

// Save writes the identity to path with restrictive permissions.
func (id *Identity) Save(path string) error {
	if id == nil || len(id.PublicKey) != ed25519.PublicKeySize || len(id.PrivateKey) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid identity")
	}
	doc := document{
		Public:  base64.StdEncoding.EncodeToString(id.PublicKey),
		Private: base64.StdEncoding.EncodeToString(id.PrivateKey),
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("serialize identity: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create identity dir: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write identity temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("commit identity: %w", err)
	}
	return nil
}

// Sign signs content with the private key.
func (id *Identity) Sign(content []byte) []byte {
	return ed25519.Sign(id.PrivateKey, content)
}

// Verify checks a signature against a public key and content.
func Verify(publicKey, content, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(publicKey), content, signature)
}

func fromKeys(pub ed25519.PublicKey, priv ed25519.PrivateKey) *Identity {
	return &Identity{
		ID:         NodeID("node:" + base64.RawURLEncoding.EncodeToString(pub)),
		PublicKey:  pub,
		PrivateKey: priv,
	}
}
