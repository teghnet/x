package transport

import (
	"net/http"

	"github.com/teghnet/x/policy"
)

// RateLimit returns a [Middleware] limiting attempts to rps per second with
// the given burst. A non-positive rps returns a no-op Middleware — the
// default is unlimited.
func RateLimit(rps float64, burst int) Middleware {
	if rps <= 0 {
		return func(next http.RoundTripper) http.RoundTripper { return next }
	}
	l := policy.NewLimiter(rps, burst)
	return func(next http.RoundTripper) http.RoundTripper {
		return RoundTripFunc(func(req *http.Request) (*http.Response, error) {
			if err := l.Wait(req.Context()); err != nil {
				closeBody(req)
				return nil, err
			}
			return next.RoundTrip(req)
		})
	}
}
