package banking

import (
	"context"
	"io"
	"iter"
	"slices"
	"strings"
	"testing"
)

type stubParser struct{}

func (stubParser) Parse(context.Context, io.Reader) iter.Seq2[*Transaction, error] {
	return func(yield func(*Transaction, error) bool) {}
}

func TestRegistry(t *testing.T) {
	Register("test_stub", func() Parser { return stubParser{} })

	p, ok := GetParser("test_stub")
	if !ok {
		t.Fatal("GetParser(test_stub): ok = false, want true")
	}
	if _, ok := p.(stubParser); !ok {
		t.Fatalf("GetParser(test_stub): got %T, want stubParser", p)
	}

	if !slices.Contains(List(), "test_stub") {
		t.Fatalf("List() = %v, want it to contain test_stub", List())
	}

	if _, ok := GetParser("does_not_exist"); ok {
		t.Fatal("GetParser(does_not_exist): ok = true, want false")
	}
}

func TestDecodeReader(t *testing.T) {
	t.Run("utf-8 passthrough", func(t *testing.T) {
		r, err := DecodeReader(strings.NewReader("héllo"), EncodingUTF8)
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "héllo" {
			t.Fatalf("got %q, want %q", got, "héllo")
		}
	})

	t.Run("windows-1250", func(t *testing.T) {
		// 0xB9 is 'ą' (LATIN SMALL LETTER A WITH OGONEK) in windows-1250.
		r, err := DecodeReader(strings.NewReader("kwota: 100 z\xb9"), EncodingWindows1250)
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if want := "kwota: 100 zą"; string(got) != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("unsupported encoding", func(t *testing.T) {
		if _, err := DecodeReader(strings.NewReader(""), "ebcdic"); err == nil {
			t.Fatal("want error for unsupported encoding, got nil")
		}
	})
}
