package index

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	got := Tokenize("Hello, World! 123 foo-bar")
	want := []string{"hello", "world", "123", "foo", "bar"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAddAndSearch(t *testing.T) {
	ix := New()
	ix.Add("d1", "the quick brown fox")
	ix.Add("d2", "the lazy dog sleeps")
	ix.Add("d3", "quick quick quick fox jumps")

	if ix.Len() != 3 {
		t.Fatalf("len = %d", ix.Len())
	}

	res := ix.Search("quick", 0)
	if len(res) != 2 {
		t.Fatalf("expected 2 hits, got %d: %+v", len(res), res)
	}
	// d3 mentions "quick" more often, so it should rank first.
	if res[0].ID != "d3" {
		t.Fatalf("expected d3 first, got %q", res[0].ID)
	}
}

func TestSearchMultiTerm(t *testing.T) {
	ix := New()
	ix.Add("a", "apple banana")
	ix.Add("b", "apple")
	res := ix.Search("apple banana", 0)
	if len(res) != 2 || res[0].ID != "a" {
		t.Fatalf("doc matching both terms should rank first: %+v", res)
	}
}

func TestSearchLimit(t *testing.T) {
	ix := New()
	ix.Add("a", "x")
	ix.Add("b", "x")
	ix.Add("c", "x")
	if got := ix.Search("x", 2); len(got) != 2 {
		t.Fatalf("limit ignored, got %d", len(got))
	}
}

func TestReindexReplaces(t *testing.T) {
	ix := New()
	ix.Add("d1", "cats")
	ix.Add("d1", "dogs") // replace
	if len(ix.Search("cats", 0)) != 0 {
		t.Fatal("old content still indexed after reindex")
	}
	if len(ix.Search("dogs", 0)) != 1 {
		t.Fatal("new content not indexed")
	}
}

func TestDelete(t *testing.T) {
	ix := New()
	ix.Add("d1", "hello")
	ix.Delete("d1")
	if ix.Len() != 0 {
		t.Fatalf("len = %d after delete", ix.Len())
	}
	if len(ix.Search("hello", 0)) != 0 {
		t.Fatal("deleted doc still searchable")
	}
}

func TestSearchEmpty(t *testing.T) {
	ix := New()
	if ix.Search("anything", 0) != nil {
		t.Fatal("empty index should return nil")
	}
	ix.Add("d", "content")
	if ix.Search("   ", 0) != nil {
		t.Fatal("empty query should return nil")
	}
}
