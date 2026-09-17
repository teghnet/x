// Package transport provides retry, backoff, and rate-limiting policies for
// clients of remote endpoints.
//
// [policy.Policy] and [policy.Policy.Do] are protocol-agnostic: they know nothing about
// HTTP and can drive retries for any request/response pair, such as a gRPC
// call, a database round trip, or an SFTP transfer. This file builds an
// [http.RoundTripper] on top of Policy for HTTP specifically, as a chain of
// [RoundTripMiddleware] values — [Retry], [RateLimit], or your own — around a base
// transport; use [policy.Policy.Do] directly to wrap other kinds of connectors.
package transport

import (
	"fmt"
	"net/http"
)

// errPrefix identifies errors originating from this package.
const errPrefix = "transport"

// New builds an [http.RoundTripper] out of base (http.DefaultTransport
// unless overridden by [WithBaseTransport]) wrapped in the middleware
// installed by [WithMiddleware], in the order given. With no middleware,
// New is a pure pass-through to base — retries, rate limiting, and anything
// else are all opt-in via [WithMiddleware].
func New(opts ...Option) *Transport {
	t := &Transport{base: http.DefaultTransport.RoundTrip}
	for _, opt := range opts {
		opt(t)
	}
	t.base = chain(t.base, t.mw)
	return t
}

// Transport is an [http.RoundTripper] built by [New]: a base transport wrapped in a chain of [RoundTripMiddleware].
type Transport struct {
	base RoundTrip
	mw   []RoundTripMiddleware
}

// Client returns an [http.Client] using t as its transport. Most callers
// want this rather than using t directly.
func (t *Transport) Client() *http.Client {
	return &http.Client{Transport: t}
}

// RoundTrip implements [http.RoundTripper]. The chain is built once, by
// [New]; this only runs it. It never modifies req: every
// [RoundTripMiddleware] in the chain is required to clone before changing anything
// on the request it receives (see [RoundTripMiddleware]), so no defensive clone is
// needed here.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.base(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errPrefix, err)
	}
	return res, nil
}

// Option configures a [Transport] built by [New].
type Option func(*Transport)

// WithBaseTransport sets the underlying transport, typically an
// authenticating transport so credentials are injected. Defaults to
// http.DefaultTransport.
func WithBaseTransport(rt RoundTrip) Option {
	return func(t *Transport) { t.base = rt }
}

// WithMiddleware installs mw into the chain, in the order given: the first
// is outermost, seeing the request first and the response last. Repeated
// calls append rather than replace.
func WithMiddleware(mw ...RoundTripMiddleware) Option {
	return func(t *Transport) { t.mw = append(t.mw, mw...) }
}
