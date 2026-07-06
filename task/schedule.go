package task

import (
	"context"
	"time"
)

// Schedule describes when a task should run. Every fires a task on a fixed
// interval; the zero Schedule never fires.
type Schedule struct {
	// Interval between runs. Must be > 0 for the schedule to fire.
	Interval time.Duration
	// Immediate runs the task once at start before waiting the first interval.
	Immediate bool
}

// Every returns a Schedule that fires on the given interval.
func Every(d time.Duration) Schedule { return Schedule{Interval: d} }

// WithImmediate returns a copy of s that also fires once at start.
func (s Schedule) WithImmediate() Schedule {
	s.Immediate = true
	return s
}

// Run executes t according to s until ctx is cancelled, invoking onResult (if
// non-nil) after each attempt with the task's result and error. It returns
// ctx.Err() when the context is done. A non-positive interval returns
// immediately with nil.
//
// Run is a helper, not a daemon: the caller decides in which goroutine it runs
// and owns the lifecycle via ctx.
func (s Schedule) Run(ctx context.Context, t Task, onResult func(Result, error)) error {
	if s.Interval <= 0 {
		return nil
	}
	fire := func() {
		res, err := t.Run(ctx)
		if onResult != nil {
			onResult(res, err)
		}
	}
	if s.Immediate {
		fire()
	}
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			fire()
		}
	}
}
