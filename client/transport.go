package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/teghnet/x/internal/backoff"
)

// retryTransport is an http.RoundTripper that applies rate limiting before each
// attempt and retries transient failures with exponential backoff and jitter.
type retryTransport struct {
	base      http.RoundTripper
	limiter   *limiter
	policy    backoff.Policy
	retryable func(*http.Response, error) bool
	sleep     func(context.Context, time.Duration) error
	jitter    func() float64
}

func randFrac() float64 { return rand.Float64() }

// defaultRetryable retries on transport errors and on 429/503/502/504 and other
// 5xx responses.
func defaultRetryable(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	switch resp.StatusCode {
	case http.StatusTooManyRequests: // 429
		return true
	}
	return resp.StatusCode >= 500
}

// RoundTrip implements http.RoundTripper.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := bufferBody(req)
	if err != nil {
		return nil, err
	}

	attempts := max(t.policy.MaxAttempts, 0)

	var resp *http.Response
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

		resp, lastErr = t.base.RoundTrip(req)

		if attempt >= attempts || !t.retryable(resp, lastErr) {
			break
		}
		// Drain and close the response body before retrying so the connection
		// can be reused.
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		delay := t.backoffDelay(resp, attempt)
		if err := t.sleep(req.Context(), delay); err != nil {
			return nil, err
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("client: request failed: %w", lastErr)
	}
	return resp, nil
}

// backoffDelay honors a Retry-After header when present, otherwise uses the
// jittered exponential policy.
func (t *retryTransport) backoffDelay(resp *http.Response, attempt int) time.Duration {
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
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
	base := t.policy.Delay(attempt)
	jittered := max(t.policy.Jitter(attempt, frac), base/2)
	return jittered
}

// bufferBody reads and returns the request body so it can be replayed on retry.
// It returns nil for bodyless requests.
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
