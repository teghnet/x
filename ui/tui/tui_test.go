package tui

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teghnet/x/store"
	"github.com/teghnet/x/task"
)

func TestCollectAndRender(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemState()
	st.Set(ctx, "alpha", []byte("1"))
	st.Set(ctx, "beta", []byte("2"))

	results := []task.Result{
		{Task: "sync", Elapsed: 1500 * time.Millisecond, Output: "done"},
	}

	snap, err := Collect(ctx, "test", results, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Keys) != 2 {
		t.Fatalf("keys = %v", snap.Keys)
	}

	var buf bytes.Buffer
	d := New(&buf)
	d.Now = func() time.Time { return time.Unix(0, 0).UTC() }
	if err := d.Render(snap); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"== test ==", "Tasks (1)", "sync", "1.5s", "State keys (2)", "alpha", "beta"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderNilState(t *testing.T) {
	snap, err := Collect(context.Background(), "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := New(&buf).Render(snap); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "ctld") {
		t.Fatalf("default title missing:\n%s", buf.String())
	}
}
