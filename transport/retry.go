// Copyright (c) 2026 Paweł Zaremba
// SPDX-License-Identifier: MIT

package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/tatenadev/x/policy"
)

// maxDiscardBody caps how much of a discarded response body [Retry] will
// read before closing it, so a large or adversarial error body can't be
// used to stall or exhaust memory on retry. Bodies longer than this may
// prevent the underlying connection from being reused, which is an
// acceptable trade-off against an unbounded read.
const maxDiscardBody = 64 * 1024

// Retrier configures the [Retry] middleware: how many times and how long to
// wait between attempts, and which responses/errors are worth retrying.
type Retrier struct {
	// Backoff controls retry timing and count.
	Backoff policy.Backoff
	// Retryable classifies whether an attempt should be retried, given the
	// request's method. Nil uses RetryableHTTP(method, Backoff.Max).
	Retryable func(method string) policy.Classifier[*http.Response]
	// OnRetry, if set, is called before each retry.
	OnRetry func(policy.RetryEvent)
}

// Retry returns a [Middleware] that retries failed attempts per r.Backoff.
// It buffers the request body (or replays it via GetBody, when available)
// so each attempt sees the original payload, and clones the request before
// every attempt so each gets its own context and body. Because Retry
// calls next more than once, everything installed after it in the chain re-runs
// on every attempt too — a downstream middleware that mints a credential
// re-mints it on each retry rather than reusing one that may have expired.
func Retry(r Retrier) Middleware {
	retryable := r.Retryable
	if retryable == nil {
		retryable = func(method string) policy.Classifier[*http.Response] {
			return RetryableHTTP(method, r.Backoff.Max)
		}
	}
	p := policy.Policy{Backoff: r.Backoff, OnRetry: r.OnRetry}
	return func(next RoundTrip) RoundTrip {
		return func(req *http.Request) (*http.Response, error) {
			maxAttempts := max(r.Backoff.MaxAttempts, 0)

			// Buffer the body only when there might be more than one
			// attempt and the request can't hand us a fresh reader itself
			// via GetBody.
			var body []byte
			switch {
			case maxAttempts <= 0 || req.Body == nil || req.Body == http.NoBody:
				// Single attempt, or nothing to replay.
			case req.GetBody != nil:
				// GetBody supplies a fresh reader per attempt; the original
				// is redundant and would otherwise never be closed.
				closeBody(req)
			default:
				b, err := io.ReadAll(req.Body)
				closeBody(req)
				if err != nil {
					return nil, fmt.Errorf("read body: %w", err)
				}
				body = b
			}

			classify := retryable(req.Method)

			attempt := func(ctx context.Context) (*http.Response, error) {
				areq := req.Clone(ctx)
				switch {
				case maxAttempts <= 0:
					// Single attempt: areq.Body already carries req's
					// original reader via the shallow copy Clone performs.
				case req.GetBody != nil:
					// req.Body was already closed above, in favor of a
					// fresh reader from GetBody on every attempt; nothing
					// new to close if GetBody itself fails.
					rc, err := req.GetBody()
					if err != nil {
						return nil, fmt.Errorf("get body: %w", err)
					}
					areq.Body = rc
				case body != nil:
					areq.Body = io.NopCloser(bytes.NewReader(body))
				}
				return next(areq)
			}

			return p.Do(req.Context(), classify, discardResponse, attempt)
		}
	}
}

// discardResponse drains and closes res.Body (bounded by maxDiscardBody) so
// the connection can be reused before a retry. It is a no-op for a nil
// response or a nil body, which a custom base [http.RoundTripper] may
// return.
func discardResponse(res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}
	io.CopyN(io.Discard, res.Body, maxDiscardBody) //nolint:errcheck // best-effort drain before Close
	res.Body.Close()
}

// idempotentMethods are HTTP methods safe to retry even when the previous
// attempt's outcome is unknown (e.g. after a connection error). Methods
// outside this set are retried only on responses that prove the request
// was never applied (429, 503 without a partial-processing risk).
var idempotentMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodPut:     true,
	http.MethodDelete:  true,
	http.MethodOptions: true,
	http.MethodTrace:   true,
}

// RetryableHTTP is the default [Classifier] for HTTP: it retries connection
// errors and 429/5xx responses for idempotent methods, and 429/503 for
// non-idempotent ones (POST, PATCH, CONNECT), on the assumption that only
// those two statuses reliably indicate the request was never processed. Any
// Retry-After the response carries is honored and clamped to max.
func RetryableHTTP(method string, max time.Duration) policy.Classifier[*http.Response] {
	return func(res *http.Response, err error) policy.Decision {
		if err != nil {
			return policy.Decision{Retry: true}
		}
		retryable := res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests
		if !retryable {
			return policy.Decision{}
		}
		if !idempotentMethods[method] && res.StatusCode != http.StatusTooManyRequests &&
			res.StatusCode != http.StatusServiceUnavailable {
			return policy.Decision{}
		}
		return policy.Decision{Retry: true, Delay: retryAfter(res, max)}
	}
}

// retryAfter parses the Retry-After header per RFC 9110 §10.2.3 — either a
// number of seconds or an HTTP-date — and clamps the result to max (when
// max > 0). It returns 0 if the header is absent or unparsable, leaving the
// caller to fall back to its own backoff.
func retryAfter(res *http.Response, max time.Duration) time.Duration {
	ra := res.Header.Get("Retry-After")
	if ra == "" {
		return 0
	}

	var delay time.Duration
	if secs, err := strconv.Atoi(ra); err == nil {
		if secs < 0 {
			return 0
		}
		delay = time.Duration(secs) * time.Second
	} else if when, err := http.ParseTime(ra); err == nil {
		delay = time.Until(when)
	} else {
		return 0
	}

	if delay <= 0 {
		return 0
	}
	if max > 0 && delay > max {
		delay = max
	}
	return delay
}
