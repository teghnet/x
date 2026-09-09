package transport

import "time"

// DefaultBackoff is a reasonable policy for HTTP retries: 200ms base,
// doubling up to 10s, with three retries.
var DefaultBackoff = Backoff{
	Base:        200 * time.Millisecond,
	Max:         10 * time.Second,
	Factor:      2.0,
	MaxAttempts: 3,
}

// Backoff describes an exponential backoff curve.
// The zero value performs no backoff (Delay and Jitter both return 0).
type Backoff struct {
	// Base is the delay before the first retry (attempt 0).
	Base time.Duration
	// Max caps the delay for any attempt. Zero means uncapped.
	Max time.Duration
	// Factor multiplies the delay each attempt (e.g. 2.0 doubles it).
	// Values less than 1 are treated as 1 (no growth), which keeps a
	// misconfigured Backoff from collapsing every retry after the first
	// down to zero.
	Factor float64
	// MaxAttempts is the number of retries after the initial try. A value
	// <= 0 means no retries.
	MaxAttempts int
}

// normalize returns b with a sane growth factor.
func (b Backoff) normalize() Backoff {
	if b.Factor < 1 {
		b.Factor = 1
	}
	return b
}

// Delay returns the base (un-jittered) delay before the given retry
// attempt, where attempt 0 is the first retry. The result is clamped to
// [0, Max].
func (b Backoff) Delay(attempt int) time.Duration {
	if attempt < 0 || b.Base <= 0 {
		return 0
	}
	b = b.normalize()
	d := float64(b.Base)
	for range attempt {
		d *= b.Factor
		if b.Max > 0 && d >= float64(b.Max) {
			return b.Max
		}
	}
	if b.Max > 0 && d > float64(b.Max) {
		return b.Max
	}
	return time.Duration(d)
}

// Jitter applies "equal jitter" to Delay(attempt): the result is uniformly
// distributed within [delay/2, delay], so callers always wait at least
// half the computed backoff. frac is expected in [0,1) and supplied by the
// caller's own randomness source, keeping this method deterministic and
// testable.
func (b Backoff) Jitter(attempt int, frac float64) time.Duration {
	delay := b.Delay(attempt)
	if delay <= 0 {
		return 0
	}
	if frac < 0 {
		frac = 0
	}
	if frac >= 1 {
		frac = 0.999999
	}
	half := delay / 2
	return half + time.Duration(float64(delay-half)*frac)
}
