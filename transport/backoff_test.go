package transport

import (
	"testing"
	"time"
)

func TestDelayGrowsAndClamps(t *testing.T) {
	p := backoff{base: 100 * time.Millisecond, max: time.Second, factor: 2}
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
		if got := p.delay(tt.attempt); got != tt.want {
			t.Errorf("Delay(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestDelayNegativeOrZeroBase(t *testing.T) {
	p := backoff{base: 100 * time.Millisecond, factor: 2}
	if p.delay(-1) != 0 {
		t.Error("negative attempt should be 0")
	}
	if (backoff{}).delay(1) != 0 {
		t.Error("zero base should be 0")
	}
}

func TestJitter(t *testing.T) {
	p := backoff{base: time.Second, max: time.Minute, factor: 2}
	if got := p.jitter(0, 0); got != 0 {
		t.Errorf("frac 0 => %v, want 0", got)
	}
	if got := p.jitter(0, 0.5); got != 500*time.Millisecond {
		t.Errorf("frac 0.5 => %v", got)
	}
	// frac >= 1 is clamped below the base delay.
	if got := p.jitter(0, 1.5); got >= time.Second {
		t.Errorf("frac >=1 not clamped: %v", got)
	}
}
