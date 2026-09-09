package transport

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

// job is a stand-in for a non-HTTP result type — Policy.Do must not care
// what T is, proving the core is protocol-agnostic.
type job struct {
	id int
	ok bool
}

var errTransient = errors.New("transient failure")

func TestPolicyDoSucceedsFirstTry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := Policy{Backoff: Backoff{Base: time.Second, MaxAttempts: 3, Factor: 2}}
		calls := 0
		classify := func(j job, err error) Decision { return Decision{} }
		result, err := p.Do(context.Background(), classify, nil, func(ctx context.Context) (job, error) {
			calls++
			return job{id: 1, ok: true}, nil
		})
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		if result.id != 1 || calls != 1 {
			t.Errorf("result=%+v calls=%d, want one successful call", result, calls)
		}
	})
}

func TestPolicyDoRetriesUntilBudgetExhausted(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := Policy{Backoff: Backoff{Base: 100 * time.Millisecond, Factor: 2, MaxAttempts: 3}}
		calls := 0
		classify := func(j job, err error) Decision { return Decision{Retry: err != nil} }
		start := time.Now()
		_, err := p.Do(context.Background(), classify, nil, func(ctx context.Context) (job, error) {
			calls++
			return job{}, errTransient
		})
		if !errors.Is(err, errTransient) {
			t.Errorf("err = %v, want errTransient", err)
		}
		if calls != 4 { // initial + 3 retries
			t.Errorf("calls = %d, want 4", calls)
		}
		// Delay(attempt) is 100ms/200ms/400ms; Jitter halves that at worst,
		// so the floor across all three retries is 50+100+200=350ms.
		if elapsed := time.Since(start); elapsed < 350*time.Millisecond {
			t.Errorf("elapsed = %v, want >= 350ms of jittered backoff", elapsed)
		}
	})
}

func TestPolicyDoStopsWhenClassifierSaysNoRetry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := Policy{Backoff: Backoff{Base: time.Second, MaxAttempts: 5, Factor: 2}}
		calls := 0
		classify := func(j job, err error) Decision { return Decision{Retry: false} }
		_, err := p.Do(context.Background(), classify, nil, func(ctx context.Context) (job, error) {
			calls++
			return job{}, errTransient
		})
		if !errors.Is(err, errTransient) {
			t.Errorf("err = %v, want errTransient", err)
		}
		if calls != 1 {
			t.Errorf("calls = %d, want 1 (classifier refused retry)", calls)
		}
	})
}

func TestPolicyDoHonorsDecisionDelayOverride(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// A large backoff base that would dominate if Decision.Delay were
		// ignored.
		p := Policy{Backoff: Backoff{Base: time.Hour, Factor: 2, MaxAttempts: 1}}
		calls := 0
		classify := func(j job, err error) Decision {
			return Decision{Retry: calls == 1, Delay: 5 * time.Second}
		}
		start := time.Now()
		_, _ = p.Do(context.Background(), classify, nil, func(ctx context.Context) (job, error) {
			calls++
			if calls == 1 {
				return job{}, errTransient
			}
			return job{ok: true}, nil
		})
		elapsed := time.Since(start)
		if elapsed != 5*time.Second {
			t.Errorf("elapsed = %v, want exactly 5s (Decision.Delay override)", elapsed)
		}
	})
}

func TestPolicyDoCallsDiscarderBeforeRetry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := Policy{Backoff: Backoff{Base: time.Millisecond, MaxAttempts: 1}}
		var discarded []int
		calls := 0
		classify := func(j job, err error) Decision { return Decision{Retry: calls < 2} }
		discard := func(j job) { discarded = append(discarded, j.id) }
		_, _ = p.Do(context.Background(), classify, discard, func(ctx context.Context) (job, error) {
			calls++
			return job{id: calls}, nil
		})
		if len(discarded) != 1 || discarded[0] != 1 {
			t.Errorf("discarded = %v, want [1] (only the thrown-away first result)", discarded)
		}
	})
}

func TestPolicyDoOnRetryFiresWithAttemptAndDelay(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var retries []RetryEvent
		p := Policy{
			Backoff: Backoff{Base: time.Second, Factor: 1, MaxAttempts: 2},
			OnRetry: func(r RetryEvent) { retries = append(retries, r) },
		}
		classify := func(j job, err error) Decision { return Decision{Retry: err != nil} }
		_, _ = p.Do(context.Background(), classify, nil, func(ctx context.Context) (job, error) {
			return job{}, errTransient
		})
		if len(retries) != 2 {
			t.Fatalf("len(retries) = %d, want 2", len(retries))
		}
		if retries[0].Attempt != 0 || retries[1].Attempt != 1 {
			t.Errorf("attempts = %d,%d, want 0,1", retries[0].Attempt, retries[1].Attempt)
		}
		for _, r := range retries {
			if !errors.Is(r.Err, errTransient) {
				t.Errorf("RetryEvent.Err = %v, want errTransient", r.Err)
			}
		}
	})
}

func TestPolicyDoRespectsContextCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := Policy{Backoff: Backoff{Base: time.Minute, MaxAttempts: 5}}
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		classify := func(j job, err error) Decision { return Decision{Retry: true} }
		done := make(chan error, 1)
		go func() {
			_, err := p.Do(ctx, classify, nil, func(ctx context.Context) (job, error) {
				calls++
				return job{}, errTransient
			})
			done <- err
		}()
		synctest.Wait()
		cancel()
		err := <-done
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled", err)
		}
		if calls != 1 {
			t.Errorf("calls = %d, want 1 (cancel during backoff sleep)", calls)
		}
	})
}
