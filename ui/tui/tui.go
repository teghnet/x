// Package tui renders a terminal dashboard of task activity and stored state.
//
// It consumes the task and store packages only; it never reaches into api/* or
// client. The renderer is deliberately dependency-free (plain text to an
// io.Writer) so it stays easy to test and embeds no terminal library. A caller
// wanting a full-screen interactive UI can build one atop Snapshot.
package tui

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/teghnet/x/store"
	"github.com/teghnet/x/task"
)

// Snapshot is the read-only data a dashboard frame renders. Callers assemble it
// from their live task results and store; the UI never pulls from services
// itself.
type Snapshot struct {
	Title   string
	Results []task.Result
	Keys    []string
}

// Dashboard renders Snapshots to an output writer.
type Dashboard struct {
	Out io.Writer
	// Now supplies the current time for the header; nil uses time.Now.
	Now func() time.Time
}

// New returns a Dashboard writing to out.
func New(out io.Writer) *Dashboard { return &Dashboard{Out: out} }

// Collect builds a Snapshot from the given results and state. It reads the
// store's keys; it does not fetch values, keeping the frame cheap.
func Collect(ctx context.Context, title string, results []task.Result, s store.State) (Snapshot, error) {
	snap := Snapshot{Title: title, Results: results}
	if s != nil {
		keys, err := s.Keys(ctx)
		if err != nil {
			return snap, fmt.Errorf("tui: collect keys: %w", err)
		}
		snap.Keys = keys
	}
	return snap, nil
}

// Render writes a single dashboard frame for snap.
func (d *Dashboard) Render(snap Snapshot) error {
	now := time.Now
	if d.Now != nil {
		now = d.Now
	}
	title := snap.Title
	if title == "" {
		title = "ctld"
	}
	if _, err := fmt.Fprintf(d.Out, "== %s == %s\n\n", title, now().Format(time.RFC3339)); err != nil {
		return err
	}

	fmt.Fprintf(d.Out, "Tasks (%d)\n", len(snap.Results))
	tw := tabwriter.NewWriter(d.Out, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "  TASK\tELAPSED\tOUTPUT")
	for _, r := range snap.Results {
		fmt.Fprintf(tw, "  %s\t%s\t%v\n", r.Task, r.Elapsed.Round(time.Millisecond), r.Output)
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(d.Out, "\nState keys (%d)\n", len(snap.Keys))
	for _, k := range snap.Keys {
		fmt.Fprintf(d.Out, "  - %s\n", k)
	}
	return nil
}
