package store

import (
	"context"
	"path/filepath"
	"testing"
)

func testState(t *testing.T, s State) {
	t.Helper()
	ctx := context.Background()

	if _, ok, err := s.Get(ctx, "missing"); err != nil || ok {
		t.Fatalf("missing key: ok=%v err=%v", ok, err)
	}
	if err := s.Set(ctx, "a", []byte("1")); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(ctx, "b", []byte("2")); err != nil {
		t.Fatal(err)
	}
	v, ok, err := s.Get(ctx, "a")
	if err != nil || !ok || string(v) != "1" {
		t.Fatalf("get a: %q ok=%v err=%v", v, ok, err)
	}
	keys, err := s.Keys(ctx)
	if err != nil || len(keys) != 2 || keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("keys = %v err=%v", keys, err)
	}
	if err := s.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := s.Get(ctx, "a"); ok {
		t.Fatal("expected a deleted")
	}
	// Deleting a missing key is not an error.
	if err := s.Delete(ctx, "a"); err != nil {
		t.Fatalf("delete missing: %v", err)
	}
}

func TestMemState(t *testing.T) { testState(t, NewMemState()) }

func TestMemStateGetReturnsCopy(t *testing.T) {
	ctx := context.Background()
	s := NewMemState()
	if err := s.Set(ctx, "k", []byte("orig")); err != nil {
		t.Fatal(err)
	}
	v, _, _ := s.Get(ctx, "k")
	v[0] = 'X' // mutate the returned slice
	again, _, _ := s.Get(ctx, "k")
	if string(again) != "orig" {
		t.Fatalf("stored value mutated: %q", again)
	}
}

func TestFileState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	s, err := OpenFileState(path)
	if err != nil {
		t.Fatal(err)
	}
	testState(t, s)
}

func TestFileStatePersistence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.json")

	s1, err := OpenFileState(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s1.Set(ctx, "k", []byte("v")); err != nil {
		t.Fatal(err)
	}

	s2, err := OpenFileState(path)
	if err != nil {
		t.Fatal(err)
	}
	v, ok, _ := s2.Get(ctx, "k")
	if !ok || string(v) != "v" {
		t.Fatalf("reopened state lost data: %q ok=%v", v, ok)
	}
}

func TestJSONHelpers(t *testing.T) {
	ctx := context.Background()
	s := NewMemState()
	type rec struct {
		N int    `json:"n"`
		S string `json:"s"`
	}
	if err := SetJSON(ctx, s, "r", rec{N: 5, S: "hi"}); err != nil {
		t.Fatal(err)
	}
	got, err := GetJSON[rec](ctx, s, "r")
	if err != nil {
		t.Fatal(err)
	}
	if got.N != 5 || got.S != "hi" {
		t.Fatalf("got %+v", got)
	}
	if _, err := GetJSON[rec](ctx, s, "nope"); err == nil {
		t.Fatal("expected ErrNotFound")
	}
}
