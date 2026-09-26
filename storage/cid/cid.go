// Package cid provides content-addressed identifiers and a local block store.
//
// Supported ID formats:
//   - "cid1:" + hex(sha256)  — Phase 3 stable format (default for Put)
//   - IPFS CIDv1 (bafkrei...) — via NewIPFS using github.com/ipfs/go-cid
//
// Why two formats: the sandbox GOPROXY (http://35.245.43.102) returned 502
// during Phase 3, so cid1 was implemented first. Official proxy.golang.org works;
// NewIPFS is available without breaking existing cid1 blocks.
package cid

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	ipfsCid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

const Prefix = "cid1:"

// New computes a Phase-3 content ID: cid1: + hex(sha256).
func New(content []byte) string {
	sum := sha256.Sum256(content)
	return Prefix + hex.EncodeToString(sum[:])
}

// NewIPFS computes an IPFS CIDv1 (raw codec, sha2-256).
func NewIPFS(content []byte) (string, error) {
	h, err := mh.Sum(content, mh.SHA2_256, -1)
	if err != nil {
		return "", fmt.Errorf("multihash: %w", err)
	}
	c := ipfsCid.NewCidV1(ipfsCid.Raw, h)
	return c.String(), nil
}

// Valid reports whether s is a known content id format.
func Valid(s string) bool {
	if strings.HasPrefix(s, Prefix) {
		h := strings.TrimPrefix(s, Prefix)
		if len(h) != 64 {
			return false
		}
		_, err := hex.DecodeString(h)
		return err == nil
	}
	// IPFS CIDv0/v1
	_, err := ipfsCid.Decode(s)
	return err == nil
}

// Store is a content-addressed block store.
type Store interface {
	Put(content []byte) (id string, err error)
	Get(id string) ([]byte, error)
	Has(id string) bool
}

// LocalStore stores blocks as files under a directory.
type LocalStore struct {
	dir string
	mu  sync.RWMutex
}

// NewLocalStore creates (if needed) a directory-backed block store.
func NewLocalStore(dir string) (*LocalStore, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create cid store: %w", err)
	}
	return &LocalStore{dir: dir}, nil
}

func (s *LocalStore) pathFor(id string) (string, error) {
	if !Valid(id) {
		return "", fmt.Errorf("invalid cid: %s", id)
	}
	var shard, name string
	if strings.HasPrefix(id, Prefix) {
		h := strings.TrimPrefix(id, Prefix)
		shard, name = h[:2], h
	} else {
		// safe filename from hash of full cid string
		sum := sha256.Sum256([]byte(id))
		name = hex.EncodeToString(sum[:])
		shard = name[:2]
	}
	return filepath.Join(s.dir, shard, name), nil
}

// Put stores content under cid1: id. Idempotent for the same bytes.
func (s *LocalStore) Put(content []byte) (string, error) {
	id := New(content)
	return id, s.putAt(id, content)
}

// PutIPFS stores content under an IPFS CIDv1 id.
func (s *LocalStore) PutIPFS(content []byte) (string, error) {
	id, err := NewIPFS(content)
	if err != nil {
		return "", err
	}
	return id, s.putAt(id, content)
}

func (s *LocalStore) putAt(id string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path, err := s.pathFor(id)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, 0600); err != nil {
		return fmt.Errorf("write block: %w", err)
	}
	return os.Rename(tmp, path)
}

// Get loads content by CID (cid1 or IPFS).
func (s *LocalStore) Get(id string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	path, err := s.pathFor(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("get block %s: %w", id, err)
	}
	// integrity: recompute expected id family
	if strings.HasPrefix(id, Prefix) {
		if New(data) != id {
			return nil, fmt.Errorf("block integrity failed for %s", id)
		}
	} else {
		expect, err := NewIPFS(data)
		if err != nil || expect != id {
			return nil, fmt.Errorf("block integrity failed for %s", id)
		}
	}
	return data, nil
}

// Has reports whether the CID is present.
func (s *LocalStore) Has(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	path, err := s.pathFor(id)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
