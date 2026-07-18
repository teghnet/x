package transport

import "time"

// defaultBackoff is a reasonable policy for HTTP retries: 200ms base,
// doubling up to 10s, with three retries.
var defaultBackoff = &backoff{
	base:        200 * time.Millisecond,
	max:         10 * time.Second,
	factor:      2.0,
	maxAttempts: 3,
}

// backoff describes an exponential backoff curve.
// The zero value is not useful;
// use defaultBackoff or construct one explicitly.
type backoff struct {
	// base is the delay before the first retry (attempt 0).
	base time.Duration
	// max caps the delay for any attempt.
	max time.Duration
	// factor multiplies the delay each attempt (e.g. 2.0 doubles it).
	factor float64
	// maxAttempts is the number of retries after the initial try. A value <= 0
	// means no retries.
	maxAttempts int
}

// delay returns the base (un-jittered) delay before the given retry attempt,
// where attempt 0 is the first retry. The result is clamped to [0, max].
func (p backoff) delay(attempt int) time.Duration {
	if attempt < 0 || p.base <= 0 {
		return 0
	}
	d := float64(p.base)
	for range attempt {
		d *= p.factor
		if p.max > 0 && d >= float64(p.max) {
			return p.max
		}
	}
	if p.max > 0 && d > float64(p.max) {
		return p.max
	}
	return time.Duration(d)
}

// jitter returns delay(attempt) scaled by frac, where frac is expected in
// [0,1). It applies "full jitter": the returned delay is uniformly within
// [0, delay]. Callers supply frac from their own randomness source so this stays
// deterministic and testable.
func (p backoff) jitter(attempt int, frac float64) time.Duration {
	base := p.delay(attempt)
	if base <= 0 {
		return 0
	}
	if frac < 0 {
		frac = 0
	}
	if frac >= 1 {
		frac = 0.999999
	}
	return time.Duration(float64(base) * frac)
}
