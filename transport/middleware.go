package transport

import (
	"net/http"
	"slices"
)

// TransportDecorator wraps an [http.RoundTripper] to add behavior around it — the
// same shape as http.Handler middleware, applied to the client side. A
// TransportDecorator must eventually call next.RoundTrip to complete the request,
// unless it fails fast (for example, a rate limiter whose context is
// already done). Per [http.RoundTripper], a TransportDecorator must not modify the
// request it receives — clone before changing anything on it — and, on any
// path that returns an error without calling next, must close the
// request's body itself (see [closeBody]), since nothing downstream will.
type TransportDecorator func(next http.RoundTripper) http.RoundTripper

// RoundTripMiddleware adapts a function to [http.RoundTripper], mirroring [http.HandlerFunc].
type RoundTripMiddleware func(*http.Request) (*http.Response, error)

// RoundTrip implements [http.RoundTripper].
func (f RoundTripMiddleware) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// closeBody closes r.Body, if any. A [TransportDecorator] that returns an error
// without calling next must call this: net/http's Client assumes the
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
func chain(base http.RoundTripper, mw []TransportDecorator) http.RoundTripper {
	rt := base
	for _, m := range slices.Backward(mw) {
		rt = m(rt)
	}
	return rt
}
