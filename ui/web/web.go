// Package web serves a small local HTML dashboard over task and store state.
//
// Assets are embedded with embed.FS so the binary is self-contained. Like the
// tui package, web consumes only task and store; it must not import api/* or
// client. The composition root supplies a snapshot function that reads live
// state.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"time"

	"github.com/teghnet/x/store"
	"github.com/teghnet/x/task"
)

//go:embed assets
var assets embed.FS

// Server exposes the dashboard as an http.Handler. Results returns the current
// task results (may be nil); State is the durable state to list keys from.
type Server struct {
	// Results returns a snapshot of recent task results. Required for the
	// results endpoint; nil yields an empty list.
	Results func() []task.Result
	// State backs the state endpoint; nil yields no keys.
	State store.State
}

// Handler returns the HTTP handler serving the dashboard UI and its JSON APIs:
//
//	GET /             embedded dashboard page
//	GET /api/results  recent task results as JSON
//	GET /api/state    stored keys as JSON
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	sub, _ := fs.Sub(assets, "assets")
	mux.Handle("GET /", http.FileServer(http.FS(sub)))
	mux.HandleFunc("GET /api/results", s.handleResults)
	mux.HandleFunc("GET /api/state", s.handleState)
	return mux
}

// resultDTO is the JSON shape for a task result; it flattens Elapsed to
// milliseconds for easy display and drops non-serializable fields.
type resultDTO struct {
	Task      string `json:"task"`
	StartedAt string `json:"started_at"`
	ElapsedMs int64  `json:"elapsed_ms"`
	Output    any    `json:"output"`
}

func (s *Server) handleResults(w http.ResponseWriter, r *http.Request) {
	var results []task.Result
	if s.Results != nil {
		results = s.Results()
	}
	dtos := make([]resultDTO, 0, len(results))
	for _, res := range results {
		dtos = append(dtos, resultDTO{
			Task:      res.Task,
			StartedAt: res.Started.Format(time.RFC3339),
			ElapsedMs: res.Elapsed.Milliseconds(),
			Output:    jsonSafe(res.Output),
		})
	}
	writeJSON(w, dtos)
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	keys := []string{}
	if s.State != nil {
		if k, err := s.State.Keys(r.Context()); err == nil {
			keys = k
		}
	}
	writeJSON(w, map[string]any{"keys": keys})
}

// jsonSafe returns v if it can be marshaled, otherwise its string form. This
// keeps arbitrary task Output payloads from breaking the endpoint.
func jsonSafe(v any) any {
	if v == nil {
		return nil
	}
	if _, err := json.Marshal(v); err != nil {
		return nil
	}
	return v
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// ListenAndServe starts the dashboard server on addr until ctx is cancelled,
// then shuts it down gracefully.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: s.Handler()}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errc:
		return err
	}
}
