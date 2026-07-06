// Package task defines the contract for units of work and how they are run.
//
// It is intentionally dependency-light: it declares interfaces (Task, Runner)
// and value types (Result) plus small adapters and helpers. All wiring of
// concrete tasks and runners lives in the composition root (cmd/ctld); this
// package never reaches into api, store or client.
package task

import (
	"context"
	"fmt"
	"time"
)

// Task is a single unit of work. Implementations must honor ctx cancellation.
// Name identifies the task for logging and scheduling and should be stable.
type Task interface {
	Name() string
	Run(ctx context.Context) (Result, error)
}

// Result is the outcome of running a Task. Output carries an optional
// task-specific payload; callers type-assert it as needed.
type Result struct {
	Task    string
	Started time.Time
	Elapsed time.Duration
	Output  any
}

// Runner executes tasks. Implementations decide sequencing, concurrency and
// error handling. Returning an error means the run as a whole failed.
type Runner interface {
	Run(ctx context.Context, tasks ...Task) ([]Result, error)
}

// Func adapts a plain function into a Task.
type Func struct {
	Label string
	Fn    func(ctx context.Context) (any, error)
}

// Name implements Task.
func (f Func) Name() string { return f.Label }

// Run implements Task, populating Result timing fields around Fn.
func (f Func) Run(ctx context.Context) (Result, error) {
	start := time.Now()
	out, err := f.Fn(ctx)
	res := Result{Task: f.Label, Started: start, Elapsed: time.Since(start), Output: out}
	if err != nil {
		return res, fmt.Errorf("task %q: %w", f.Label, err)
	}
	return res, nil
}
