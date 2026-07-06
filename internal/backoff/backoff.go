// Package backoff computes exponential backoff delays with optional jitter.
//
// It is an internal helper shared by client (HTTP retries) and other packages
// that must wait between attempts. It is pure and deterministic given a jitter
// source, which makes it easy to test.
package backoff

import "time"

// Policy describes an exponential backoff curve. The zero value is not useful;
// use Default or construct one explicitly.
type Policy struct {
	// Base is the delay before the first retry (attempt 0).
	Base time.Duration
	// Max caps the delay for any attempt.
	Max time.Duration
	// Factor multiplies the delay each attempt (e.g. 2.0 doubles it).
	Factor float64
	// MaxAttempts is the number of retries after the initial try. A value <= 0
	// means no retries.
	MaxAttempts int
}

// Default is a reasonable policy for HTTP retries: 200ms base, doubling up to
// 10s, with three retries.
func Default() Policy {
	return Policy{
		Base:        200 * time.Millisecond,
		Max:         10 * time.Second,
		Factor:      2.0,
		MaxAttempts: 3,
	}
}

// Delay returns the base (un-jittered) delay before the given retry attempt,
// where attempt 0 is the first retry. The result is clamped to [0, Max].
func (p Policy) Delay(attempt int) time.Duration {
	if attempt < 0 || p.Base <= 0 {
		return 0
	}
	d := float64(p.Base)
	for range attempt {
		d *= p.Factor
		if p.Max > 0 && d >= float64(p.Max) {
			return p.Max
		}
	}
	if p.Max > 0 && d > float64(p.Max) {
		return p.Max
	}
	return time.Duration(d)
}

// Jitter returns Delay(attempt) scaled by frac, where frac is expected in
// [0,1). It applies "full jitter": the returned delay is uniformly within
// [0, Delay]. Callers supply frac from their own randomness source so this stays
// deterministic and testable.
func (p Policy) Jitter(attempt int, frac float64) time.Duration {
	base := p.Delay(attempt)
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
