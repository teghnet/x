package json

import (
	"errors"
	"strings"
	"testing"
)

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    point
		wantErr bool
	}{
		{name: "ok", in: `{"x":1,"y":2}`, want: point{1, 2}},
		{name: "partial", in: `{"x":3}`, want: point{X: 3}},
		{name: "extra field ignored", in: `{"x":1,"y":2,"z":3}`, want: point{1, 2}},
		{name: "empty object", in: `{}`, want: point{}},
		{name: "malformed", in: `{"x":`, wantErr: true},
		{name: "wrong type", in: `{"x":"nope"}`, wantErr: true},
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
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalStrict(t *testing.T) {
	if _, err := UnmarshalStrict[point]([]byte(`{"x":1,"z":9}`)); err == nil {
		t.Fatal("expected strict decode to reject unknown field")
	}
	got, err := UnmarshalStrict[point]([]byte(`{"x":1,"y":2}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != (point{1, 2}) {
		t.Fatalf("got %+v", got)
	}
}

func TestDecode(t *testing.T) {
	got, err := Decode[[]int](strings.NewReader(`[1,2,3]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("got %v", got)
	}
}

func TestStream(t *testing.T) {
	var sum int
	err := Stream[point](strings.NewReader(`{"x":1}{"x":2} {"x":3}`), func(p point) error {
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
	err := Stream[point](strings.NewReader(`{"x":1}{"x":2}`), func(p point) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want sentinel", err)
	}
}
