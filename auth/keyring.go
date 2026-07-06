package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ErrKeyNotFound is returned by Keyring implementations when a key is absent.
var ErrKeyNotFound = errors.New("auth: key not found")

// Keyring is secret storage keyed by (service, account). Implementations must be
// safe for concurrent use.
//
// The interface deliberately abstracts over the backing store so a real OS
// keychain can be swapped in later without touching callers; the bundled
// implementations are an in-memory ring and a file-backed ring.
type Keyring interface {
	Get(service, account string) (string, error)
	Set(service, account, secret string) error
	Delete(service, account string) error
}

func ringKey(service, account string) string { return service + "\x00" + account }

// MemKeyring is an in-memory Keyring, useful for tests and ephemeral processes.
type MemKeyring struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewMemKeyring returns an empty in-memory keyring.
func NewMemKeyring() *MemKeyring {
	return &MemKeyring{m: make(map[string]string)}
}

// Get implements Keyring.
func (k *MemKeyring) Get(service, account string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	v, ok := k.m[ringKey(service, account)]
	if !ok {
		return "", fmt.Errorf("%s/%s: %w", service, account, ErrKeyNotFound)
	}
	return v, nil
}

// Set implements Keyring.
func (k *MemKeyring) Set(service, account, secret string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.m[ringKey(service, account)] = secret
	return nil
}

// Delete implements Keyring.
func (k *MemKeyring) Delete(service, account string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.m, ringKey(service, account))
	return nil
}

// FileKeyring persists secrets to a JSON file with owner-only permissions. It is
// a pragmatic fallback where no OS keychain is available; it does not encrypt at
// rest, so protect the file's location accordingly.
type FileKeyring struct {
	path string
	mu   sync.RWMutex
	m    map[string]string
}

// OpenFileKeyring loads a file-backed keyring, creating an empty one if the file
// does not exist.
func OpenFileKeyring(path string) (*FileKeyring, error) {
	k := &FileKeyring{path: path, m: make(map[string]string)}
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return k, nil
	case err != nil:
		return nil, fmt.Errorf("auth: open keyring %s: %w", path, err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &k.m); err != nil {
			return nil, fmt.Errorf("auth: parse keyring %s: %w", path, err)
		}
	}
	return k, nil
}

// Get implements Keyring.
func (k *FileKeyring) Get(service, account string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	v, ok := k.m[ringKey(service, account)]
	if !ok {
		return "", fmt.Errorf("%s/%s: %w", service, account, ErrKeyNotFound)
	}
	return v, nil
}

// Set implements Keyring.
func (k *FileKeyring) Set(service, account, secret string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.m[ringKey(service, account)] = secret
	return k.flushLocked()
}

// Delete implements Keyring.
func (k *FileKeyring) Delete(service, account string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.m, ringKey(service, account))
	return k.flushLocked()
}

func (k *FileKeyring) flushLocked() error {
	data, err := json.Marshal(k.m)
	if err != nil {
		return fmt.Errorf("auth: encode keyring: %w", err)
	}
	if dir := filepath.Dir(k.path); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("auth: mkdir keyring: %w", err)
		}
	}
	if err := os.WriteFile(k.path, data, 0o600); err != nil {
		return fmt.Errorf("auth: write keyring: %w", err)
	}
	return nil
}
