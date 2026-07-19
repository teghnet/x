package csv

import (
	"errors"
	"strings"
	"testing"
)

type point struct {
	X int `csv:"x"`
	Y int `csv:"y"`
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []point
		wantErr bool
	}{
		{name: "ok", in: "x,y\n1,2\n3,4\n", want: []point{{1, 2}, {3, 4}}},
		{name: "partial", in: "x,y\n3,\n", want: []point{{X: 3}}},
		{name: "extra column ignored", in: "x,y,z\n1,2,3\n", want: []point{{1, 2}}},
		{name: "reordered header", in: "y,x\n2,1\n", want: []point{{1, 2}}},
		{name: "header only", in: "x,y\n", want: nil},
		{name: "empty input", in: "", want: nil},
		{name: "malformed", in: "x,y\n\"1,2\n", wantErr: true},
		{name: "wrong type", in: "x,y\nnope,2\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unmarshal[point]([]byte(tt.in))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestDecodeUnknownFieldNotStruct(t *testing.T) {
	if _, err := Unmarshal[int]([]byte("x\n1\n")); err == nil {
		t.Fatal("expected error decoding into non-struct type")
	}
}

func TestStream(t *testing.T) {
	var sum int
	err := Stream(strings.NewReader("x,y\n1,0\n2,0\n3,0\n"), func(p point) error {
		sum += p.X
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum != 6 {
		t.Fatalf("sum = %d, want 6", sum)
	}
}

func TestStreamCallbackError(t *testing.T) {
	sentinel := errors.New("stop")
	err := Stream(strings.NewReader("x,y\n1,0\n2,0\n"), func(p point) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want sentinel", err)
	}
}

func TestAll(t *testing.T) {
	var sum int
	for p, err := range Iter[point](strings.NewReader("x,y\n1,0\n2,0\n3,0\n")) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		sum += p.X
	}
	if sum != 6 {
		t.Fatalf("sum = %d, want 6", sum)
	}
}

func TestAllError(t *testing.T) {
	var got []point
	var gotErr error
	for p, err := range Iter[point](strings.NewReader("x,y\nnope,2\n")) {
		if err != nil {
			gotErr = err
			break
		}
		got = append(got, p)
	}
	if gotErr == nil {
		t.Fatal("expected error for wrong type")
	}
	if len(got) != 0 {
		t.Fatalf("got %+v, want none", got)
	}
}

func TestAllBreak(t *testing.T) {
	var seen int
	for range Iter[point](strings.NewReader("x,y\n1,0\n2,0\n3,0\n")) {
		seen++
		break
	}
	if seen != 1 {
		t.Fatalf("seen = %d, want 1", seen)
	}
}

func TestFieldNameFallback(t *testing.T) {
	type row struct {
		Name string
		Age  int
	}
	got, err := Unmarshal[row]([]byte("Name,Age\nAda,36\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := row{Name: "Ada", Age: 36}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestTagDash(t *testing.T) {
	type row struct {
		Name string `csv:"-"`
		Age  int    `csv:"age"`
	}
	got, err := Unmarshal[row]([]byte("Name,age\nAda,36\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := row{Age: 36}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestTagPosition(t *testing.T) {
	// A blank header (e.g. a leading ID column some exports leave unnamed)
	// can't be matched by name, so it's targeted by 0-based index instead.
	type row struct {
		ID   string `csv:",0"`
		Name string `csv:"name"`
	}
	got, err := Unmarshal[row]([]byte(",name\n42,Ada\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := row{ID: "42", Name: "Ada"}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestTagPositionMultipleBlankHeaders(t *testing.T) {
	// Two blank-header columns in the same row must be distinguishable by
	// position; an untagged field never matches either.
	type row struct {
		First string `csv:",0"`
		Mid   string `csv:"mid"`
		Last  string `csv:",2"`
	}
	got, err := Unmarshal[row]([]byte(",mid,\na,b,c\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := row{First: "a", Mid: "b", Last: "c"}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestTagPositionInvalid(t *testing.T) {
	type row struct {
		Bad string `csv:",oops"`
	}
	if _, err := Unmarshal[row]([]byte(",\nx\n")); err == nil {
		t.Fatal("expected error for non-integer csv tag position")
	}
}
