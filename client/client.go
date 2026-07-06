// Package client is the shared HTTP transport used by all api/* packages.
//
// A Client bundles three cross-cutting concerns behind a single http.Client:
//
//   - authentication, injected as a base http.RoundTripper (see package auth),
//   - rate limiting, via a token-bucket limiter,
//   - retries with exponential backoff and jitter on transient failures.
//
// api/* packages depend on Client for transport and never construct these
// behaviors themselves.
package client

import (
	"net/http"
	"time"

	"github.com/teghnet/x/internal/backoff"
)

// Client is a configured HTTP client. Construct it with New; the zero value is
// not usable. It is safe for concurrent use.
type Client struct {
	hc *http.Client
}

// Option configures a Client.
type Option func(*config)

type config struct {
	base      http.RoundTripper
	limiter   *limiter
	retry     backoff.Policy
	retryable func(*http.Response, error) bool
	timeout   time.Duration
}

// WithBaseTransport sets the underlying transport, typically an *auth.Transport
// so that credentials are injected. Defaults to http.DefaultTransport.
func WithBaseTransport(rt http.RoundTripper) Option {
	return func(c *config) { c.base = rt }
}

// WithRateLimit limits requests to rps per second with the given burst. A
// non-positive rps disables rate limiting.
func WithRateLimit(rps float64, burst int) Option {
	return func(c *config) {
		if rps > 0 {
			c.limiter = newLimiter(rps, burst)
		}
	}
}

// WithRetry sets the backoff policy for retries. Use backoff.Policy{} to disable.
func WithRetry(p backoff.Policy) Option {
	return func(c *config) { c.retry = p }
}

// WithRetryable overrides the predicate deciding whether a response/error should
// be retried. The default retries connection errors and 429/5xx responses.
func WithRetryable(fn func(*http.Response, error) bool) Option {
	return func(c *config) { c.retryable = fn }
}

// WithTimeout sets a per-request timeout on the underlying http.Client. This
// bounds the whole call including retries.
func WithTimeout(d time.Duration) Option {
	return func(c *config) { c.timeout = d }
}

// New builds a Client from the given options.
func New(opts ...Option) *Client {
	cfg := &config{
		base:      http.DefaultTransport,
		retry:     backoff.Default(),
		retryable: defaultRetryable,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	rt := &retryTransport{
		base:      cfg.base,
		limiter:   cfg.limiter,
		policy:    cfg.retry,
		retryable: cfg.retryable,
		sleep:     sleepCtx,
		jitter:    randFrac,
	}
	return &Client{hc: &http.Client{Transport: rt, Timeout: cfg.timeout}}
}

// HTTPClient returns the underlying *http.Client for callers that need it
// (e.g. to pass to libraries expecting one). Mutating its Transport is
// unsupported.
func (c *Client) HTTPClient() *http.Client { return c.hc }

// Do sends an HTTP request and returns the response, applying rate limiting and
// retries. It mirrors http.Client.Do.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.hc.Do(req)
}
