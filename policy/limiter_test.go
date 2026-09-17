package policy

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

func TestLimiterNilDisabled(t *testing.T) {
	var l *Limiter
	if err := l.Wait(context.Background()); err != nil {
		t.Errorf("nil limiter should not block: %v", err)
	}
}

func TestLimiterNonPositiveRateDisabled(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(0, 5)
		start := time.Now()
		for range 100 {
			if err := l.Wait(context.Background()); err != nil {
				t.Fatalf("Wait: %v", err)
			}
		}
		if elapsed := time.Since(start); elapsed != 0 {
			t.Errorf("disabled limiter should never wait, elapsed %v", elapsed)
		}
	})
}

func TestLimiterBurstThenThrottles(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(1, 3) // 1 token/sec, burst of 3
		start := time.Now()

		// The initial burst should drain instantly.
		for range 3 {
			if err := l.Wait(context.Background()); err != nil {
				t.Fatalf("Wait: %v", err)
			}
		}
		if elapsed := time.Since(start); elapsed != 0 {
			t.Errorf("burst should not wait, elapsed %v", elapsed)
		}

		// The 4th call must wait roughly 1s for the next token.
		if err := l.Wait(context.Background()); err != nil {
			t.Fatalf("Wait: %v", err)
		}
		if elapsed := time.Since(start); elapsed < time.Second {
			t.Errorf("expected to wait ~1s for refill, elapsed %v", elapsed)
		}
	})
}

func TestLimiterRefillsOverFakeTime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(10, 1) // 10 tokens/sec, burst of 1
		ctx := context.Background()

		if err := l.Wait(ctx); err != nil {
			t.Fatalf("Wait: %v", err)
		}
		start := time.Now()
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("Wait: %v", err)
		}
		elapsed := time.Since(start)
		if elapsed < 100*time.Millisecond {
			t.Errorf("expected ~100ms between tokens at 10/s, elapsed %v", elapsed)
		}
	})
}

func TestLimiterCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(0.001, 1) // effectively never refills within the test
		ctx := context.Background()
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("first Wait (uses burst token): %v", err)
		}

		cctx, cancel := context.WithCancel(ctx)
		done := make(chan error, 1)
		go func() { done <- l.Wait(cctx) }()

		synctest.Wait() // let the goroutine block on its timer
		cancel()

		if err := <-done; err != context.Canceled {
			t.Errorf("Wait after cancel = %v, want context.Canceled", err)
		}
	})
}

func TestLimiterConcurrentWaitersRace(t *testing.T) {
	// Deliberately outside synctest: this exercises the mutex under real
	// goroutine scheduling for -race, not fake-clock semantics.
	l := NewLimiter(1000, 10)
	ctx := context.Background()

	done := make(chan struct{})
	for range 20 {
		go func() {
			defer func() { done <- struct{}{} }()
			for range 5 {
				if err := l.Wait(ctx); err != nil {
					t.Errorf("Wait: %v", err)
				}
			}
		}()
	}
	for range 20 {
		<-done
	}
}
