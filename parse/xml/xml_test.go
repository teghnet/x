package xml

import (
	"errors"
	"strings"
	"testing"
)

type feed struct {
	Title string `xml:"title"`
	Items []item `xml:"item"`
}

type item struct {
	Name string `xml:"name"`
	ID   int    `xml:"id,attr"`
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    feed
		wantErr bool
	}{
		{
			name: "ok",
			in:   `<feed><title>news</title><item id="1"><name>a</name></item></feed>`,
			want: feed{Title: "news", Items: []item{{Name: "a", ID: 1}}},
		},
		{
			name: "empty",
			in:   `<feed></feed>`,
			want: feed{},
		},
		{
			name:    "malformed",
			in:      `<feed><title>oops`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unmarshal[feed]([]byte(tt.in))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Title != tt.want.Title || len(got.Items) != len(tt.want.Items) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestStream(t *testing.T) {
	doc := `<feed><item id="1"><name>a</name></item><item id="2"><name>b</name></item></feed>`
	var ids []int
	err := Stream[item](strings.NewReader(doc), "item", func(it item) error {
		ids = append(ids, it.ID)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("ids = %v", ids)
	}
}

func TestStreamCallbackError(t *testing.T) {
	sentinel := errors.New("stop")
	err := Stream[item](strings.NewReader(`<feed><item id="1"/></feed>`), "item", func(item) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("got %v, want sentinel", err)
	}
}

func TestAll(t *testing.T) {
	doc := `<feed><item id="1"><name>a</name></item><item id="2"><name>b</name></item></feed>`
	var ids []int
	for it, err := range All[item](strings.NewReader(doc), "item") {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		ids = append(ids, it.ID)
	}
	if len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("ids = %v", ids)
	}
}

func TestAllBreak(t *testing.T) {
	doc := `<feed><item id="1"/><item id="2"/><item id="3"/></feed>`
	var seen int
	for range All[item](strings.NewReader(doc), "item") {
		seen++
		break
	}
	if seen != 1 {
		t.Fatalf("seen = %d, want 1", seen)
	}
}
