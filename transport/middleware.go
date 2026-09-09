package transport

import (
	"fmt"
	"net/http"
)

// Middleware wraps an [http.RoundTripper] to add behavior around it — the
// same shape as http.Handler middleware, applied to the client side. A
// Middleware must eventually call next.RoundTrip to complete the request,
// unless it fails fast (for example, a rate limiter whose context is
// already done).
type Middleware func(next http.RoundTripper) http.RoundTripper

// RoundTripFunc adapts a function to [http.RoundTripper], mirroring
// [http.HandlerFunc].
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements [http.RoundTripper].
func (f RoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// MutateRequest returns a [Middleware] that calls fn on the request before
// passing it to next.
//
// fn may modify the request in place: every request entering a [Transport]
// built by [New] is already a fresh clone ([Transport.RoundTrip] clones
// once per call; [Retry] clones again before each attempt), so mutating it
// here never reaches the caller's original request or leaks into another
// attempt. Using MutateRequest to wrap an arbitrary [http.RoundTripper]
// outside of a Transport chain does not carry that guarantee — the caller
// would need to clone first.
func MutateRequest(fn func(*http.Request) error) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(r *http.Request) (*http.Response, error) {
			if err := fn(r); err != nil {
				return nil, fmt.Errorf("mutate request: %w", err)
			}
			return next.RoundTrip(r)
		})
	}
}

// chain composes mw around base. mw[0] ends up outermost: it sees the
// request first and the response last, and — for a middleware like Retry
// that calls next more than once — every middleware after it re-runs on
// each of those calls.
func chain(base http.RoundTripper, mw []Middleware) http.RoundTripper {
	rt := base
	for i := len(mw) - 1; i >= 0; i-- {
		rt = mw[i](rt)
	}
	return rt
}
