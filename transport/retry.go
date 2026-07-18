package transport

import "net/http"

// defaultRetryable retries on config errors
// and on 429/503/502/504
// and other 5xx responses.
func defaultRetryable(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	switch resp.StatusCode {
	case http.StatusTooManyRequests: // 429
		return true
	}
	return resp.StatusCode >= 500
}
