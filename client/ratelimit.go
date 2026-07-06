package client

import (
	"context"
	"sync"
	"time"
)

// limiter is a token-bucket rate limiter. Tokens refill continuously at a fixed
// rate up to a burst capacity. It avoids a background goroutine by computing
// available tokens lazily from elapsed time.
type limiter struct {
	mu     sync.Mutex
	rate   float64   // tokens per second
	burst  float64   // maximum tokens
	tokens float64   // current tokens
	last   time.Time // last refill time
	now    func() time.Time
	sleep  func(context.Context, time.Duration) error
}

// newLimiter builds a limiter allowing rate requests per second with the given
// burst. A rate <= 0 disables limiting (wait always returns immediately).
func newLimiter(rate float64, burst int) *limiter {
	b := float64(burst)
	if b < 1 {
		b = 1
	}
	return &limiter{
		rate:   rate,
		burst:  b,
		tokens: b,
		now:    time.Now,
		sleep:  sleepCtx,
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
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

// wait blocks until a token is available or ctx is cancelled.
func (l *limiter) wait(ctx context.Context) error {
	if l == nil || l.rate <= 0 {
		return ctx.Err()
	}
	for {
		l.mu.Lock()
		now := l.now()
		if l.last.IsZero() {
			l.last = now
		}
		// Refill.
		elapsed := now.Sub(l.last).Seconds()
		if elapsed > 0 {
			l.tokens = min(l.burst, l.tokens+elapsed*l.rate)
			l.last = now
		}
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		// Time until the next whole token.
		deficit := 1 - l.tokens
		wait := time.Duration(deficit / l.rate * float64(time.Second))
		l.mu.Unlock()

		if err := l.sleep(ctx, wait); err != nil {
			return err
		}
	}
}
