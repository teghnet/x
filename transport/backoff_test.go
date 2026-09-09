package transport

import (
	"testing"
	"time"
)

func TestDelayGrowsAndClamps(t *testing.T) {
	b := Backoff{Base: 100 * time.Millisecond, Max: time.Second, Factor: 2}
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, time.Second}, // clamped
		{10, time.Second},
	}
	for _, tt := range tests {
		if got := b.Delay(tt.attempt); got != tt.want {
			t.Errorf("Delay(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestDelayNegativeOrZeroBase(t *testing.T) {
	b := Backoff{Base: 100 * time.Millisecond, Factor: 2}
	if b.Delay(-1) != 0 {
		t.Error("negative attempt should be 0")
	}
	if (Backoff{}).Delay(1) != 0 {
		t.Error("zero base should be 0")
	}
}

// TestDelayFactorBelowOneNormalizes ensures a misconfigured Factor doesn't
// collapse retries to zero delay: Factor < 1 is treated as 1 (constant
// delay), not as decay toward zero.
func TestDelayFactorBelowOneNormalizes(t *testing.T) {
	b := Backoff{Base: 100 * time.Millisecond, Max: time.Second, Factor: 0}
	for attempt := range 5 {
		if got := b.Delay(attempt); got != 100*time.Millisecond {
			t.Errorf("Delay(%d) with zero factor = %v, want 100ms (normalized)", attempt, got)
		}
	}
}

func TestJitter(t *testing.T) {
	b := Backoff{Base: time.Second, Max: time.Minute, Factor: 2}
	// Equal jitter: result is always within [delay/2, delay].
	if got := b.Jitter(0, 0); got != 500*time.Millisecond {
		t.Errorf("frac 0 => %v, want 500ms (the floor)", got)
	}
	if got := b.Jitter(0, 1); got < 999*time.Millisecond || got > time.Second {
		t.Errorf("frac ~1 => %v, want ~1s", got)
	}
	if got := b.Jitter(0, 0.5); got != 750*time.Millisecond {
		t.Errorf("frac 0.5 => %v, want 750ms (midpoint)", got)
	}
	// frac >= 1 is clamped just below the top of the range.
	if got := b.Jitter(0, 1.5); got >= time.Second {
		t.Errorf("frac >=1 not clamped: %v", got)
	}
	if got := b.Jitter(0, 1.5); got < 500*time.Millisecond {
		t.Errorf("frac >=1 fell below the floor: %v", got)
	}
}

func TestJitterZeroDelay(t *testing.T) {
	if got := (Backoff{}).Jitter(0, 0.5); got != 0 {
		t.Errorf("zero backoff should jitter to 0, got %v", got)
	}
}
