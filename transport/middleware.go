package transport

import (
	"net/http"
	"slices"
)

// RoundTripFn is one hop of the chain: a function that completes a round
// trip. It satisfies [http.RoundTripper], so it also serves as a base
// transport via [WithBaseTransport].
type RoundTripFn func(*http.Request) (*http.Response, error)

// RoundTrip implements [http.RoundTripper].
func (f RoundTripFn) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// RoundTripMiddleware wraps a single round trip: it receives the request and
// the next [RoundTripFn] in the chain. A RoundTripMiddleware must eventually
// call next to complete the request, unless it fails fast (for example, a
// rate limiter whose context is already done). Per [http.RoundTripper] it
// must not modify req — clone before changing anything on it — and, on any
// path that returns an error without calling next, must close the request's
// body itself (see [closeBody]), since nothing downstream will.
type RoundTripMiddleware func(req *http.Request, next RoundTripFn) (*http.Response, error)

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
func chain(base RoundTripFn, mw []RoundTripMiddleware) RoundTripFn {
	rtf := base
	for _, m := range slices.Backward(mw) {
		next := rtf // captured per iteration; rtf is reassigned below
		rtf = func(req *http.Request) (*http.Response, error) {
			return m(req, next)
		}
	}
	return rtf
}
