// Package transport provides retry, backoff, and rate-limiting policies for
// clients of remote endpoints.
//
// [Policy] and [Policy.Do] are protocol-agnostic: they know nothing about
// HTTP and can drive retries for any request/response pair, such as a gRPC
// call, a database round trip, or an SFTP transfer. This file builds an
// [http.RoundTripper] on top of Policy for HTTP specifically; use
// [Policy.Do] directly to wrap other kinds of connectors.
package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// errPrefix identifies errors originating from this package.
const errPrefix = "transport"

// maxDiscardBody caps how much of a discarded response body Transport will
// read before closing it, so a large or adversarial error body can't be
// used to stall or exhaust memory on retry. Bodies longer than this may
// prevent the underlying connection from being reused, which is an
// acceptable trade-off against an unbounded read.
const maxDiscardBody = 64 * 1024

// New builds an [http.RoundTripper] that layers rate limiting, retries with
// backoff, and request mutation over base (http.DefaultTransport unless
// overridden by [WithBaseTransport]). By default it makes no retries and
// applies no rate limit; use [WithRetry] and [WithRateLimit] to enable
// them.
func New(opts ...Option) *Transport {
	t := &Transport{base: http.DefaultTransport}
	t.policy.Backoff = DefaultBackoff
	t.retryable = func(method string) Classifier[*http.Response] {
		return RetryableHTTP(method, t.policy.Backoff.Max)
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Transport is an [http.RoundTripper] that retries failed requests with
// backoff, optionally rate-limits attempts, and can mutate each outgoing
// request (for example, to inject credentials). Build one with [New].
type Transport struct {
	base    http.RoundTripper
	mutator RequestMutator

	policy Policy

	// retryable derives the retry [Classifier] for a request's method, so
	// the default can apply method-aware idempotency rules ([RetryableHTTP])
	// while [WithRetryable] can still install a fixed classifier.
	retryable func(method string) Classifier[*http.Response]
}

// Client returns an [http.Client] using t as its transport. Most callers
// want this rather than using t directly.
func (t *Transport) Client() *http.Client {
	return &http.Client{Transport: t}
}

// RoundTrip implements [http.RoundTripper]. It never modifies req; each
// attempt operates on a clone.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	maxAttempts := max(t.policy.Backoff.MaxAttempts, 0)

	// Buffer the body only when there might be more than one attempt and
	// the request can't hand us a fresh reader itself via GetBody.
	var body []byte
	if maxAttempts > 0 && req.GetBody == nil && req.Body != nil && req.Body != http.NoBody {
		b, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: read body: %w", errPrefix, err)
		}
		body = b
	}

	classify := t.retryable(req.Method)

	attempt := func(ctx context.Context) (*http.Response, error) {
		areq := req.Clone(ctx)
		switch {
		case maxAttempts <= 0:
			// Single attempt: areq.Body already carries req's original
			// reader via the shallow copy Clone performs.
		case req.GetBody != nil:
			rc, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("%s: get body: %w", errPrefix, err)
			}
			areq.Body = rc
		case body != nil:
			areq.Body = io.NopCloser(bytes.NewReader(body))
		}
		if t.mutator != nil {
			if err := t.mutator.ApplyTo(areq); err != nil {
				return nil, fmt.Errorf("%s: mutate request: %w", errPrefix, err)
			}
		}
		return t.base.RoundTrip(areq)
	}

	res, err := t.policy.Do(req.Context(), classify, discardResponse, attempt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errPrefix, err)
	}
	return res, nil
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

// RequestMutator applies a change to an outgoing request, such as setting
// an Authorization header. It is invoked on a fresh clone before every
// attempt — including retries — so a mutator backed by a refreshable
// credential (an OAuth2 token source, say) is re-consulted each time.
type RequestMutator interface {
	ApplyTo(*http.Request) error
}

// Option configures a [Transport] built by [New].
type Option func(*Transport)

// WithBaseTransport sets the underlying transport, typically an
// authenticating transport so credentials are injected. Defaults to
// http.DefaultTransport.
func WithBaseTransport(rt http.RoundTripper) Option {
	return func(t *Transport) { t.base = rt }
}

// WithRateLimit limits attempts to rps per second with the given burst. A
// non-positive rps disables rate limiting (the default).
func WithRateLimit(rps float64, burst int) Option {
	return func(t *Transport) {
		if rps > 0 {
			t.policy.Limiter = NewLimiter(rps, burst)
		} else {
			t.policy.Limiter = nil
		}
	}
}

// WithRetry sets the backoff policy for retries. Use Backoff{} to disable
// retries entirely.
func WithRetry(b Backoff) Option {
	return func(t *Transport) { t.policy.Backoff = b }
}

// WithRetryable overrides the classifier deciding whether a response/error
// should be retried, replacing [RetryableHTTP]'s method-aware default with
// a single classifier used for every request regardless of method.
func WithRetryable(classify Classifier[*http.Response]) Option {
	return func(t *Transport) {
		t.retryable = func(string) Classifier[*http.Response] { return classify }
	}
}

// WithOnRetry sets a hook called before each retry, for logging (e.g. via
// log/slog) or metrics. It must not block.
func WithOnRetry(fn func(Retry)) Option {
	return func(t *Transport) { t.policy.OnRetry = fn }
}

// WithRequestMutator sets the mutator applied to a clone of each outgoing
// request before it is sent.
func WithRequestMutator(m RequestMutator) Option {
	return func(t *Transport) { t.mutator = m }
}
