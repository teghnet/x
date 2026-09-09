package transport

import (
	"context"
	"math/rand/v2"
	"time"
)

// Decision reports what should happen after one attempt.
type Decision struct {
	// Retry asks Policy.Do for another attempt, budget permitting.
	Retry bool
	// Delay, when >0, overrides the computed backoff for this retry (for
	// example, a server-provided Retry-After). Ignored when Retry is false.
	Delay time.Duration
}

// Classifier inspects one attempt's result and error and decides whether
// Policy.Do should retry.
type Classifier[T any] func(T, error) Decision

// Discarder releases a result that Policy.Do is about to throw away before
// retrying — for example, draining and closing an HTTP response body so
// the underlying connection can be reused. May be nil.
type Discarder[T any] func(T)

// Retry describes a retry Policy.Do is about to make.
type Retry struct {
	// Attempt is the number of the attempt that just failed, starting at 0.
	Attempt int
	// Delay is how long Policy.Do will wait before the next attempt.
	Delay time.Duration
	// Err is the error from the failed attempt, if any.
	Err error
}

// Policy governs a sequence of attempts against a remote endpoint: how many
// times to retry, how long to wait between attempts, and how fast attempts
// may be made. The zero value makes a single attempt with no rate limit.
type Policy struct {
	// Backoff controls retry timing and count.
	Backoff Backoff
	// Limiter caps the rate of attempts. Nil means unlimited.
	Limiter *Limiter
	// OnRetry, if set, is called before each retry — a natural hook for
	// logging (e.g. via log/slog) or metrics. It must not block.
	OnRetry func(Retry)
}

// Do runs fn, retrying according to p until classify reports no more
// retries are wanted, the attempt budget is spent, or ctx is done. discard
// is called on each result Do is about to discard before retrying; it may
// be nil.
//
// Do is a generic method: it declares its own type parameter T in addition
// to Policy's (none), so one Policy value can drive attempts against any
// result type without Policy itself being generic — a method may declare
// type parameters independently of its receiver's [Go 1.27].
func (p Policy) Do[T any](
	ctx context.Context,
	classify Classifier[T],
	discard Discarder[T],
	fn func(context.Context) (T, error),
) (T, error) {
	attempts := max(p.Backoff.MaxAttempts, 0)

	var result T
	var lastErr error
	for attempt := 0; ; attempt++ {
		if err := ctx.Err(); err != nil {
			var zero T
			return zero, err
		}
		if err := p.Limiter.Wait(ctx); err != nil {
			var zero T
			return zero, err
		}

		result, lastErr = fn(ctx)
		decision := classify(result, lastErr)

		if attempt >= attempts || !decision.Retry {
			return result, lastErr
		}
		if discard != nil {
			discard(result)
		}

		delay := decision.Delay
		if delay <= 0 {
			delay = p.Backoff.Jitter(attempt, rand.Float64())
		}
		if p.OnRetry != nil {
			p.OnRetry(Retry{Attempt: attempt, Delay: delay, Err: lastErr})
		}
		if err := sleep(ctx, delay); err != nil {
			var zero T
			return zero, err
		}
	}
}

// sleep blocks for d or until ctx is done, whichever comes first.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
