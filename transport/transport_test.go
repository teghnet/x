package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/teghnet/x/policy"
)

// newTestClient builds an *http.Client that sends requests to srv through a
// Transport wired to srv's in-memory network, with mw installed as its
// middleware chain.
func newTestClient(srv *httptest.Server, mw ...Middleware) *http.Client {
	return New(
		WithBaseTransport(srv.Client().Transport.RoundTrip),
		WithMiddleware(mw...),
	).Client()
}

func TestTransportRetriesOn5xxThenSucceeds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) <= 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))

		client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: 10 * time.Millisecond, Factor: 2, MaxAttempts: 3}}))
		res, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want 200", res.StatusCode)
		}
		if got := calls.Load(); got != 3 {
			t.Errorf("calls = %d, want 3", got)
		}
	})
}

func TestTransportNoRetryOn4xx(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusNotFound)
		}))

		client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: 10 * time.Millisecond, Factor: 2, MaxAttempts: 3}}))
		res, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Errorf("StatusCode = %d, want 404", res.StatusCode)
		}
		if got := calls.Load(); got != 1 {
			t.Errorf("calls = %d, want 1 (4xx is not retried)", got)
		}
	})
}

func TestTransportExhaustsRetriesAndReturnsLastError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))

		client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: 10 * time.Millisecond, Factor: 2, MaxAttempts: 2}}))
		res, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("StatusCode = %d, want 503 (last attempt's response, not an error)", res.StatusCode)
		}
		if got := calls.Load(); got != 3 { // initial + 2 retries
			t.Errorf("calls = %d, want 3", got)
		}
	})
}

func TestTransportBodyReplayedViaGetBody(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const payload = "hello, retried world"
		var calls atomic.Int32
		var gotBodies []string
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			gotBodies = append(gotBodies, string(b))
			// 503, not 500: POST is non-idempotent, and the default
			// classifier only retries a non-idempotent method on 429/503.
			if calls.Add(1) <= 2 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))

		client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: time.Millisecond, Factor: 1, MaxAttempts: 3}}))
		// http.NewRequest sets GetBody automatically for a strings.Reader,
		// exercising the GetBody-per-attempt path.
		req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(payload))
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		defer res.Body.Close()

		if len(gotBodies) != 3 {
			t.Fatalf("server saw %d requests, want 3", len(gotBodies))
		}
		for i, b := range gotBodies {
			if b != payload {
				t.Errorf("attempt %d body = %q, want %q", i, b, payload)
			}
		}
	})
}

func TestTransportBodyReplayedViaBuffering(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const payload = "no GetBody here"
		var gotBodies []string
		var calls atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			gotBodies = append(gotBodies, string(b))
			if calls.Add(1) <= 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))

		client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: time.Millisecond, Factor: 1, MaxAttempts: 3}}))
		req, err := http.NewRequest(http.MethodPost, srv.URL, io.NopCloser(bytes.NewReader([]byte(payload))))
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.GetBody = nil // force the buffering path (finding 7)

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		defer res.Body.Close()

		if len(gotBodies) != 2 {
			t.Fatalf("server saw %d requests, want 2", len(gotBodies))
		}
		for i, b := range gotBodies {
			if b != payload {
				t.Errorf("attempt %d body = %q, want %q", i, b, payload)
			}
		}
	})
}

func TestTransportNonIdempotentMethodOnlyRetries429And503(t *testing.T) {
	t.Run("500 is not retried", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: time.Millisecond, Factor: 1, MaxAttempts: 3}}))
			res, err := client.Post(srv.URL, "text/plain", nil)
			if err != nil {
				t.Fatalf("Post: %v", err)
			}
			res.Body.Close()
			if got := calls.Load(); got != 1 {
				t.Errorf("calls = %d, want 1 (POST + 500 is not retried)", got)
			}
		})
	})

	t.Run("503 is retried", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) <= 1 {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: time.Millisecond, Factor: 1, MaxAttempts: 3}}))
			res, err := client.Post(srv.URL, "text/plain", nil)
			if err != nil {
				t.Fatalf("Post: %v", err)
			}
			res.Body.Close()
			if got := calls.Load(); got != 2 {
				t.Errorf("calls = %d, want 2 (POST + 503 is retried)", got)
			}
		})
	})
}

func TestTransportRetryAfterSecondsIsHonoredAndClamped(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) <= 1 {
				w.Header().Set("Retry-After", "1000000") // far beyond Max
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))

		var retries []policy.RetryEvent
		client := newTestClient(srv, Retry(Retrier{
			Backoff: policy.Backoff{Base: time.Millisecond, Factor: 1, Max: 2 * time.Second, MaxAttempts: 1},
			OnRetry: func(r policy.RetryEvent) { retries = append(retries, r) },
		}))
		start := time.Now()
		res, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		res.Body.Close()

		if len(retries) != 1 {
			t.Fatalf("len(retries) = %d, want 1", len(retries))
		}
		if retries[0].Delay != 2*time.Second {
			t.Errorf("RetryEvent.Delay = %v, want 2s (clamped to Backoff.Max)", retries[0].Delay)
		}
		if elapsed := time.Since(start); elapsed != 2*time.Second {
			t.Errorf("elapsed = %v, want exactly 2s", elapsed)
		}
	})
}

func TestTransportRetryAfterHTTPDateIsHonored(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) <= 1 {
				w.Header().Set("Retry-After", time.Now().Add(3*time.Second).UTC().Format(http.TimeFormat))
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))

		client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: time.Millisecond, Factor: 1, MaxAttempts: 1}}))
		start := time.Now()
		res, err := client.Get(srv.URL)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		res.Body.Close()

		// Allow a little slack: the HTTP-date has second resolution.
		if elapsed := time.Since(start); elapsed < 2*time.Second || elapsed > 3*time.Second {
			t.Errorf("elapsed = %v, want ~3s (Retry-After HTTP-date)", elapsed)
		}
	})
}

func TestTransportContextCancellationDuringBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		client := newTestClient(srv, Retry(Retrier{Backoff: policy.Backoff{Base: time.Minute, Factor: 1, MaxAttempts: 5}}))
		ctx, cancel := context.WithCancel(context.Background())
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		if err != nil {
			t.Fatalf("NewRequestWithContext: %v", err)
		}

		done := make(chan error, 1)
		go func() {
			_, err := client.Do(req)
			done <- err
		}()
		synctest.Wait()
		cancel()

		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want a context.Canceled-wrapping error", err)
		}
	})
}

func TestDiscardResponseDrainsAndCloses(t *testing.T) {
	body := &trackedBody{Reader: strings.NewReader(strings.Repeat("x", 1000))}
	discardResponse(&http.Response{Body: body})
	if !body.closed {
		t.Error("expected response body to be closed")
	}
}

func TestDiscardResponseNilSafe(t *testing.T) {
	discardResponse(nil)
	discardResponse(&http.Response{Body: nil})
}

func TestDiscardResponseBoundsRead(t *testing.T) {
	body := &trackedBody{Reader: strings.NewReader(strings.Repeat("x", maxDiscardBody*2))}
	discardResponse(&http.Response{Body: body})
	if !body.closed {
		t.Error("expected response body to be closed even when larger than the cap")
	}
}

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}
