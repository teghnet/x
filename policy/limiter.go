package policy

import (
	"context"
	"sync"
	"time"
)

// NewLimiter builds a Limiter allowing rate requests per second with the
// given burst. A rate <= 0 disables limiting (Wait always returns
// immediately, modulo ctx).
func NewLimiter(rate float64, burst int) *Limiter {
	b := float64(burst)
	if b < 1 {
		b = 1
	}
	return &Limiter{
		rate:   rate,
		burst:  b,
		tokens: b,
	}
}

// Limiter is a token-bucket rate limiter. Tokens refill continuously at a
// fixed rate up to a burst capacity. It avoids a background goroutine by
// computing available tokens lazily from elapsed time.
type Limiter struct {
	mu     sync.Mutex
	rate   float64   // tokens per second
	burst  float64   // maximum tokens
	tokens float64   // current tokens
	last   time.Time // last refill time
}

// Wait blocks until a token is available or ctx is done.
func (l *Limiter) Wait(ctx context.Context) error {
	if l == nil || l.rate <= 0 {
		return ctx.Err()
	}
	for {
		l.mu.Lock()
		now := time.Now()
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
		// Release the mutex before sleeping: mutex contention alone is not
		// durably blocking under testing/synctest, so holding this lock
		// while sleeping would prevent a synctest bubble's fake clock from
		// advancing.
		l.mu.Unlock()

		if err := sleep(ctx, wait); err != nil {
			return err
		}
	}
}
