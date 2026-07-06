package task

import (
	"context"
	"errors"
)

// Sequential runs tasks one after another in the order given. If StopOnError is
// true it aborts at the first failing task and returns that error together with
// the results gathered so far; otherwise it runs every task and returns the
// joined errors.
type Sequential struct {
	StopOnError bool
}

// Run implements Runner.
func (s Sequential) Run(ctx context.Context, tasks ...Task) ([]Result, error) {
	results := make([]Result, 0, len(tasks))
	var errs []error
	for _, t := range tasks {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		res, err := t.Run(ctx)
		results = append(results, res)
		if err != nil {
			if s.StopOnError {
				return results, err
			}
			errs = append(errs, err)
		}
	}
	return results, errors.Join(errs...)
}

// Concurrent runs tasks in parallel, bounded by Limit simultaneous tasks. A
// Limit <= 0 means unlimited. Results are returned in the same order as the
// input tasks. All tasks run to completion; their errors are joined.
type Concurrent struct {
	Limit int
}

// Run implements Runner.
func (c Concurrent) Run(ctx context.Context, tasks ...Task) ([]Result, error) {
	results := make([]Result, len(tasks))
	errs := make([]error, len(tasks))

	limit := c.Limit
	if limit <= 0 {
		limit = len(tasks)
	}
	if limit == 0 {
		return nil, nil
	}
	sem := make(chan struct{}, limit)
	done := make(chan int, len(tasks))

	for i, t := range tasks {
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; done <- i }()
			results[i], errs[i] = t.Run(ctx)
		}()
	}
	for range tasks {
		<-done
	}
	return results, errors.Join(errs...)
}
