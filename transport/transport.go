package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"
)

func New(opts ...Option) http.RoundTripper {
	t := &transport{
		base:      http.DefaultTransport,
		mutator:   nil,
		limiter:   defaultLimiter,
		backoff:   defaultBackoff,
		retryable: defaultRetryable,
		sleep:     defaultSleeper,
		jitter:    rand.Float64,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

type transport struct {
	base    http.RoundTripper
	mutator RequestMutator

	limiter *limiter
	backoff *backoff

	retryable func(*http.Response, error) bool
	sleep     func(context.Context, time.Duration) error
	jitter    func() float64
}

// RoundTrip implements http.RoundTripper.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.mutator != nil {
		req = req.Clone(req.Context())
		err := t.mutator.ApplyTo(req)
		if err != nil {
			return nil, err
		}
	}

	body, err := bufferBody(req)
	if err != nil {
		return nil, err
	}

	attempts := max(t.backoff.maxAttempts, 0)

	var res *http.Response
	var lastErr error
	for attempt := 0; ; attempt++ {
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		// Rate-limit each attempt.
		if err := t.limiter.wait(req.Context()); err != nil {
			return nil, err
		}
		// Restore the body for this attempt.
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		res, lastErr = t.base.RoundTrip(req)

		if attempt >= attempts || !t.retryable(res, lastErr) {
			break
		}
		// Drain and close the response body before retrying so the connection
		// can be reused.
		if res != nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}
		delay := t.backoffDelay(res, attempt)
		if err := t.sleep(req.Context(), delay); err != nil {
			return nil, err
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("client: request failed: %w", lastErr)
	}
	return res, nil
}

// backoffDelay honors a Retry-After header when present, otherwise uses the
// jittered exponential backoff.
func (t *transport) backoffDelay(res *http.Response, attempt int) time.Duration {
	if res != nil {
		if ra := res.Header.Get("Retry-After"); ra != "" {
			if secs, err := time.ParseDuration(ra + "s"); err == nil && secs > 0 {
				return secs
			}
		}
	}
	frac := 0.5
	if t.jitter != nil {
		frac = t.jitter()
	}
	// Full jitter around the exponential delay, with a floor of half the base
	// so we always wait a little.
	base := t.backoff.delay(attempt)
	jittered := max(t.backoff.jitter(attempt, frac), base/2)
	return jittered
}

type RequestMutator interface {
	ApplyTo(*http.Request) error
}

// bufferBody reads and returns the request body so it can be replayed on retry.
// It returns nil for bodiless requests.
func bufferBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}
	if req.GetBody != nil {
		// Prefer the caller-provided replay source.
		rc, err := req.GetBody()
		if err != nil {
			return nil, fmt.Errorf("client: get body: %w", err)
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	data, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("client: read body: %w", err)
	}
	return data, nil
}

type Option func(*transport)

// WithBaseTransport sets the underlying transport, typically an *auth.Transport
// so that credentials are injected. Defaults to http.DefaultTransport.
func WithBaseTransport(rt http.RoundTripper) Option {
	return func(t *transport) { t.base = rt }
}

// WithRateLimit limits requests to rps per second with the given burst. A
// non-positive rps disables rate limiting.
func WithRateLimit(rps float64, burst int) Option {
	return func(t *transport) {
		if rps > 0 {
			t.limiter = newLimiter(rps, burst)
		}
	}
}

// WithRetry sets the backoff policy for retries. Use backoff.Policy{} to disable.
func WithRetry(p *backoff) Option {
	if p == nil {
		panic("nil backoff")
	}
	return func(t *transport) { t.backoff = p }
}

// WithRetryable overrides the predicate deciding whether a response/error should
// be retried. The default retries connection errors and 429/5xx responses.
func WithRetryable(fn func(*http.Response, error) bool) Option {
	return func(t *transport) { t.retryable = fn }
}

func WithRequestMutator(h RequestMutator) Option {
	return func(t *transport) { t.mutator = h }
}
