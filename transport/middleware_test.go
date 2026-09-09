package transport

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"testing/synctest"
	"time"
)

// okRoundTripper always succeeds with a bare 200 response.
type okRoundTripper struct{}

func (okRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func TestChainFirstMiddlewareIsOutermost(t *testing.T) {
	var order []string
	mark := func(name string) Middleware {
		return func(next http.RoundTripper) http.RoundTripper {
			return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				order = append(order, name+":enter")
				res, err := next.RoundTrip(req)
				order = append(order, name+":exit")
				return res, err
			})
		}
	}

	rt := chain(okRoundTripper{}, []Middleware{mark("a"), mark("b"), mark("c")})
	if _, err := rt.RoundTrip(&http.Request{}); err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}

	want := []string{"a:enter", "b:enter", "c:enter", "c:exit", "b:exit", "a:exit"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q (full: %v)", i, order[i], want[i], order)
		}
	}
}

func TestChainEmptyIsBasePassthrough(t *testing.T) {
	rt := chain(okRoundTripper{}, nil)
	res, err := rt.RoundTrip(&http.Request{})
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", res.StatusCode)
	}
}

func TestMutateRequestAppliesBeforeNext(t *testing.T) {
	var gotHeader string
	next := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotHeader = req.Header.Get("X-Set")
		return &http.Response{StatusCode: http.StatusOK}, nil
	})
	mw := MutateRequest(func(r *http.Request) error {
		r.Header.Set("X-Set", "yes")
		return nil
	})
	req := &http.Request{Header: http.Header{}}
	if _, err := mw(next).RoundTrip(req); err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if gotHeader != "yes" {
		t.Errorf("header seen by next = %q, want %q", gotHeader, "yes")
	}
}

func TestMutateRequestShortCircuitsOnError(t *testing.T) {
	errMutate := errors.New("mutate boom")
	called := false
	next := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: http.StatusOK}, nil
	})
	mw := MutateRequest(func(r *http.Request) error { return errMutate })

	_, err := mw(next).RoundTrip(&http.Request{Header: http.Header{}})
	if !errors.Is(err, errMutate) {
		t.Errorf("err = %v, want to wrap errMutate", err)
	}
	if called {
		t.Error("next was called despite the mutator failing")
	}
}

func TestRateLimitNonPositiveIsNoOp(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		mw := RateLimit(0, 5)
		next := okRoundTripper{}
		start := time.Now()
		for range 50 {
			if _, err := mw(next).RoundTrip(&http.Request{}); err != nil {
				t.Fatalf("RoundTrip: %v", err)
			}
		}
		if elapsed := time.Since(start); elapsed != 0 {
			t.Errorf("non-positive rps should never wait, elapsed %v", elapsed)
		}
	})
}

func TestRateLimitThrottlesAcrossCalls(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		mw := RateLimit(1, 1) // 1/sec, burst 1
		rt := mw(okRoundTripper{})
		ctx := context.Background()

		start := time.Now()
		for range 3 {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
			if _, err := rt.RoundTrip(req); err != nil {
				t.Fatalf("RoundTrip: %v", err)
			}
		}
		if elapsed := time.Since(start); elapsed < 2*time.Second {
			t.Errorf("elapsed = %v, want >= 2s across 3 calls at 1/s", elapsed)
		}
	})
}

func TestRateLimitRespectsContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		mw := RateLimit(0.001, 1) // effectively never refills within the test
		rt := mw(okRoundTripper{})

		ctx := context.Background()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
		if _, err := rt.RoundTrip(req); err != nil {
			t.Fatalf("first RoundTrip (uses burst token): %v", err)
		}

		cctx, cancel := context.WithCancel(ctx)
		req2, _ := http.NewRequestWithContext(cctx, http.MethodGet, "http://example.com", nil)
		done := make(chan error, 1)
		go func() { _, err := rt.RoundTrip(req2); done <- err }()

		synctest.Wait()
		cancel()

		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled", err)
		}
	})
}
