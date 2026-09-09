package transport

import (
	"net/http"
	"strconv"
	"time"
)

// idempotentMethods are HTTP methods safe to retry even when the previous
// attempt's outcome is unknown (e.g. after a connection error). Methods
// outside this set are retried only on responses that prove the request
// was never applied (429, 503 without a partial-processing risk).
var idempotentMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodPut:     true,
	http.MethodDelete:  true,
	http.MethodOptions: true,
	http.MethodTrace:   true,
}

// RetryableHTTP is the default [Classifier] for HTTP: it retries connection
// errors and 429/5xx responses for idempotent methods, and 429/503 for
// non-idempotent ones (POST, PATCH, CONNECT), on the assumption that only
// those two statuses reliably indicate the request was never processed. Any
// Retry-After the response carries is honored and clamped to max.
func RetryableHTTP(method string, max time.Duration) Classifier[*http.Response] {
	return func(res *http.Response, err error) Decision {
		if err != nil {
			return Decision{Retry: true}
		}
		retryable := res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests
		if !retryable {
			return Decision{}
		}
		if !idempotentMethods[method] && res.StatusCode != http.StatusTooManyRequests &&
			res.StatusCode != http.StatusServiceUnavailable {
			return Decision{}
		}
		return Decision{Retry: true, Delay: retryAfter(res, max)}
	}
}

// retryAfter parses the Retry-After header per RFC 9110 §10.2.3 — either a
// number of seconds or an HTTP-date — and clamps the result to max (when
// max > 0). It returns 0 if the header is absent or unparsable, leaving the
// caller to fall back to its own backoff.
func retryAfter(res *http.Response, max time.Duration) time.Duration {
	ra := res.Header.Get("Retry-After")
	if ra == "" {
		return 0
	}

	var delay time.Duration
	if secs, err := strconv.Atoi(ra); err == nil {
		if secs < 0 {
			return 0
		}
		delay = time.Duration(secs) * time.Second
	} else if when, err := http.ParseTime(ra); err == nil {
		delay = time.Until(when)
	} else {
		return 0
	}

	if delay <= 0 {
		return 0
	}
	if max > 0 && delay > max {
		delay = max
	}
	return delay
}
