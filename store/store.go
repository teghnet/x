// Package store holds local application state.
//
// It separates three concerns behind distinct interfaces so callers depend only
// on what they use:
//
//   - State (this file): durable key/value state.
//   - Cache (cache.go): ephemeral values with a time-to-live.
//   - index (subpackage): a queryable full-text index.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

// ErrNotFound is returned by helpers that treat a missing key as an error.
// The core State.Get instead reports absence via its boolean result.
var ErrNotFound = errors.New("store: key not found")

// State is durable key/value storage. Implementations must be safe for
// concurrent use. A missing key is reported by the ok result, not an error.
type State interface {
	// Get returns the value for key. ok is false if the key is absent.
	Get(ctx context.Context, key string) (value []byte, ok bool, err error)
	// Set stores value under key, overwriting any existing value.
	Set(ctx context.Context, key string, value []byte) error
	// Delete removes key. Deleting an absent key is not an error.
	Delete(ctx context.Context, key string) error
	// Keys returns all stored keys in lexical order.
	Keys(ctx context.Context) ([]string, error)
}

// MemState is an in-memory State. The zero value is not usable; call NewMemState.
type MemState struct {
	mu sync.RWMutex
	m  map[string][]byte
}

// NewMemState returns an empty in-memory State.
func NewMemState() *MemState {
	return &MemState{m: make(map[string][]byte)}
}

// Get implements State.
func (s *MemState) Get(_ context.Context, key string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	if !ok {
		return nil, false, nil
	}
	return slices.Clone(v), true, nil
}

// Set implements State.
func (s *MemState) Set(_ context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = slices.Clone(value)
	return nil
}

// Delete implements State.
func (s *MemState) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
	return nil
}

// Keys implements State.
func (s *MemState) Keys(_ context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Sorted(maps.Keys(s.m)), nil
}

// FileState is a State persisted to a single JSON file on disk. It keeps the
// whole map in memory and rewrites the file atomically on every mutation, which
// suits small-to-moderate state (config, tokens, cursors) rather than large or
// high-churn data.
type FileState struct {
	path string
	mu   sync.RWMutex
	m    map[string][]byte
}

// OpenFileState loads state from path, creating an empty store if the file does
// not yet exist. Parent directories are created as needed.
func OpenFileState(path string) (*FileState, error) {
	s := &FileState{path: path, m: make(map[string][]byte)}
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return s, nil
	case err != nil:
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &s.m); err != nil {
			return nil, fmt.Errorf("store: parse %s: %w", path, err)
		}
	}
	return s, nil
}

// Get implements State.
func (s *FileState) Get(_ context.Context, key string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	if !ok {
		return nil, false, nil
	}
	return slices.Clone(v), true, nil
}

// Set implements State.
func (s *FileState) Set(_ context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = slices.Clone(value)
	return s.flushLocked()
}

// Delete implements State.
func (s *FileState) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[key]; !ok {
		return nil
	}
	delete(s.m, key)
	return s.flushLocked()
}

// Keys implements State.
func (s *FileState) Keys(_ context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Sorted(maps.Keys(s.m)), nil
}

// flushLocked writes the map to disk atomically. The caller must hold s.mu.
func (s *FileState) flushLocked() error {
	data, err := json.Marshal(s.m)
	if err != nil {
		return fmt.Errorf("store: encode: %w", err)
	}
	if dir := filepath.Dir(s.path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("store: mkdir: %w", err)
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".store-*")
	if err != nil {
		return fmt.Errorf("store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("store: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("store: close: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("store: rename: %w", err)
	}
	return nil
}

// GetJSON is a typed convenience over State: it fetches key and decodes the
// stored bytes as JSON into a value of type T. It returns ErrNotFound if the
// key is absent.
func GetJSON[T any](ctx context.Context, s State, key string) (T, error) {
	var v T
	data, ok, err := s.Get(ctx, key)
	if err != nil {
		return v, err
	}
	if !ok {
		return v, fmt.Errorf("%q: %w", key, ErrNotFound)
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return v, fmt.Errorf("store: decode %q: %w", key, err)
	}
	return v, nil
}

// SetJSON encodes value as JSON and stores it under key.
func SetJSON[T any](ctx context.Context, s State, key string, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("store: encode %q: %w", key, err)
	}
	return s.Set(ctx, key, data)
}
