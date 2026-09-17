package transport

import (
	"net/http"

	"github.com/teghnet/x/policy"
)

// RateLimit returns a [RoundTripMiddleware] limiting attempts to rps per second with
// the given burst. A non-positive rps returns a no-op RoundTripMiddleware — the
// default is unlimited.
func RateLimit(rps float64, burst int) RoundTripMiddleware {
	if rps <= 0 {
		return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
			return next.RoundTrip(req)
		}
	}
	l := policy.NewLimiter(rps, burst)
	return func(req *http.Request, next http.RoundTripper) (*http.Response, error) {
		if err := l.Wait(req.Context()); err != nil {
			closeBody(req)
			return nil, err
		}
		return next.RoundTrip(req)
	}
}
