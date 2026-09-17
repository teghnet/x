package transport

import (
	"context"
	"errors"
	"io"
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
	mark := func(name string) RoundTripMiddleware {
		return func(next RoundTrip) RoundTrip {
			return func(req *http.Request) (*http.Response, error) {
				order = append(order, name+":enter")
				res, err := next(req)
				order = append(order, name+":exit")
				return res, err
			}
		}
	}

	rt := chain(okRoundTripper{}.RoundTrip, []RoundTripMiddleware{mark("a"), mark("b"), mark("c")})
	if _, err := rt(&http.Request{}); err != nil {
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
	rt := chain(okRoundTripper{}.RoundTrip, nil)
	res, err := rt(&http.Request{})
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", res.StatusCode)
	}
}

// closeTrackingBody is an [io.ReadCloser] that records whether Close was
// called.
type closeTrackingBody struct {
	closed bool
}

func (b *closeTrackingBody) Read([]byte) (int, error) { return 0, io.EOF }

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return nil
}

func TestRateLimitNonPositiveIsNoOp(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rt := RateLimit(0, 5)(okRoundTripper{}.RoundTrip)
		start := time.Now()
		for range 50 {
			if _, err := rt(&http.Request{}); err != nil {
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
		// One application of the middleware: the limiter lives in the
		// resulting RoundTrip, so its tokens are shared across calls.
		rt := RateLimit(1, 1)(okRoundTripper{}.RoundTrip) // 1/sec, burst 1
		ctx := context.Background()

		start := time.Now()
		for range 3 {
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
			if _, err := rt(req); err != nil {
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
		// 0.001/sec: effectively never refills within the test.
		rt := RateLimit(0.001, 1)(okRoundTripper{}.RoundTrip)

		ctx := context.Background()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
		if _, err := rt(req); err != nil {
			t.Fatalf("first RoundTrip (uses burst token): %v", err)
		}

		cctx, cancel := context.WithCancel(ctx)
		req2, _ := http.NewRequestWithContext(cctx, http.MethodGet, "http://example.com", nil)
		done := make(chan error, 1)
		go func() { _, err := rt(req2); done <- err }()

		synctest.Wait()
		cancel()

		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled", err)
		}
	})
}

func TestRateLimitClosesBodyOnContextError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// 0.001/sec: effectively never refills within the test.
		rt := RateLimit(0.001, 1)(okRoundTripper{}.RoundTrip)

		ctx := context.Background()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
		if _, err := rt(req); err != nil {
			t.Fatalf("first RoundTrip (uses burst token): %v", err)
		}

		cctx, cancel := context.WithCancel(ctx)
		body := &closeTrackingBody{}
		req2, _ := http.NewRequestWithContext(cctx, http.MethodGet, "http://example.com", body)
		done := make(chan error, 1)
		go func() { _, err := rt(req2); done <- err }()

		synctest.Wait()
		cancel()
		<-done

		if !body.closed {
			t.Error("request body was not closed when Wait returned an error")
		}
	})
}
