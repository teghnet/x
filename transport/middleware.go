package transport

import (
	"fmt"
	"net/http"
	"slices"
)

// Middleware wraps an [http.RoundTripper] to add behavior around it — the
// same shape as http.Handler middleware, applied to the client side. A
// Middleware must eventually call next.RoundTrip to complete the request,
// unless it fails fast (for example, a rate limiter whose context is
// already done). Per [http.RoundTripper], a Middleware must not modify the
// request it receives — clone before changing anything on it — and, on any
// path that returns an error without calling next, must close the
// request's body itself (see [closeBody]), since nothing downstream will.
type Middleware func(next http.RoundTripper) http.RoundTripper

// RoundTripFunc adapts a function to [http.RoundTripper], mirroring
// [http.HandlerFunc].
type RoundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements [http.RoundTripper].
func (f RoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// MutateRequest returns a [Middleware] that calls fn on a clone of the
// request before passing it to next. fn receives a private clone — the
// [http.Request] r itself passed to MutateRequest's RoundTripper is never
// modified — so fn is free to change headers, URL, or anything else without
// regard for who else holds a reference to the original request.
func MutateRequest(fn func(*http.Request) error) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(r *http.Request) (*http.Response, error) {
			r2 := r.Clone(r.Context())
			if err := fn(r2); err != nil {
				closeBody(r)
				return nil, fmt.Errorf("mutate request: %w", err)
			}
			return next.RoundTrip(r2)
		})
	}
}

// closeBody closes r.Body, if any. A [Middleware] that returns an error
// without calling next must call this: net/http's Client assumes the
// RoundTripper it invoked already closed the request body on any error
// path, and will not close it itself.
func closeBody(r *http.Request) {
	if r.Body != nil {
		r.Body.Close() //nolint:errcheck // best-effort close on the error path
	}
}

// chain composes mw around base. mw[0] ends up outermost: it sees the
// request first and the response last, and — for a middleware like Retry
// that calls next more than once — every middleware after it re-runs on
// each of those calls.
func chain(base http.RoundTripper, mw []Middleware) http.RoundTripper {
	rt := base
	for _, m := range slices.Backward(mw) {
		rt = m(rt)
	}
	return rt
}
