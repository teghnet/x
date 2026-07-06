package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teghnet/x/store"
	"github.com/teghnet/x/task"
)

func TestResultsEndpoint(t *testing.T) {
	srv := &Server{
		Results: func() []task.Result {
			return []task.Result{{Task: "sync", Elapsed: 2 * time.Second, Output: "ok"}}
		},
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/results")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got []resultDTO
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Task != "sync" || got[0].ElapsedMs != 2000 {
		t.Fatalf("got %+v", got)
	}
}

func TestStateEndpoint(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemState()
	st.Set(ctx, "k1", []byte("v"))

	srv := &Server{State: st}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/state")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got["keys"]) != 1 || got["keys"][0] != "k1" {
		t.Fatalf("keys = %v", got["keys"])
	}
}

func TestIndexServed(t *testing.T) {
	srv := &Server{}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	if !strings.Contains(string(buf[:n]), "<!doctype html>") {
		t.Fatalf("index not served: %q", buf[:n])
	}
}

func TestListenAndServeShutdown(t *testing.T) {
	srv := &Server{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe(ctx, "127.0.0.1:0") }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("shutdown error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not shut down")
	}
}
