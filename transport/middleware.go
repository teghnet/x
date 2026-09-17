package transport

import (
	"net/http"
	"slices"
)

// RoundTripMiddleware wraps a single round trip: it receives the request and
// the next [http.RoundTripper] in the chain. A RoundTripMiddleware must
// eventually call next.RoundTrip to complete the request, unless it fails
// fast (for example, a rate limiter whose context is already done). Per
// [http.RoundTripper] it must not modify req — clone before changing
// anything on it — and, on any path that returns an error without calling
// next, must close the request's body itself (see [closeBody]), since
// nothing downstream will.
type RoundTripMiddleware func(req *http.Request, next http.RoundTripper) (*http.Response, error)

// roundTripperFunc adapts a function to [http.RoundTripper], mirroring
// [http.HandlerFunc]. It exists only so chain can bind a middleware to its
// next hop; middleware authors never need it.
type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements [http.RoundTripper].
func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// closeBody closes r.Body, if any. A [RoundTripMiddleware] that returns an
// error without calling next must call this: net/http's Client assumes the
// RoundTripper it invoked already closed the request body on any error
// path, and will not close it itself.
func closeBody(r *http.Request) {
	if r.Body != nil {
		_ = r.Body.Close()
	}
}

// chain composes mw around base.
// mw[0] ends up outermost: it sees the request first and the response last,
// and — for a middleware like Retry that calls next more than once — every middleware after it
// re-runs on each of those calls.
func chain(base http.RoundTripper, mw []RoundTripMiddleware) http.RoundTripper {
	rt := base
	for _, m := range slices.Backward(mw) {
		next := rt // captured per iteration; rt is reassigned below
		rt = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return m(req, next)
		})
	}
	return rt
}
