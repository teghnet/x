package store

import (
	"errors"
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := NewCache[string, int]()
	c.Set("a", 1, 0)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("got %v ok=%v", v, ok)
	}
	if _, ok := c.Get("missing"); ok {
		t.Fatal("expected miss")
	}
}

func TestCacheExpiry(t *testing.T) {
	c := NewCache[string, int]()
	now := time.Unix(0, 0)
	c.now = func() time.Time { return now }

	c.Set("a", 1, time.Minute)
	if _, ok := c.Get("a"); !ok {
		t.Fatal("should be present before expiry")
	}
	now = now.Add(2 * time.Minute)
	if _, ok := c.Get("a"); ok {
		t.Fatal("should be expired")
	}
	if c.Len() != 0 {
		t.Fatalf("expired entry not purged, len=%d", c.Len())
	}
}

func TestCacheGetOrLoad(t *testing.T) {
	c := NewCache[string, int]()
	calls := 0
	load := func() (int, error) { calls++; return 42, nil }

	for range 3 {
		v, err := c.GetOrLoad("k", time.Minute, load)
		if err != nil || v != 42 {
			t.Fatalf("got %v err=%v", v, err)
		}
	}
	if calls != 1 {
		t.Fatalf("load called %d times, want 1", calls)
	}
}

func TestCacheGetOrLoadError(t *testing.T) {
	c := NewCache[string, int]()
	boom := errors.New("boom")
	if _, err := c.GetOrLoad("k", time.Minute, func() (int, error) { return 0, boom }); !errors.Is(err, boom) {
		t.Fatalf("got %v", err)
	}
	if _, ok := c.Get("k"); ok {
		t.Fatal("failed load must not populate cache")
	}
}

func TestCacheJanitor(t *testing.T) {
	c := NewCache[string, int]()
	c.StartJanitor(time.Millisecond)
	defer c.Close()
	c.Set("a", 1, time.Minute)
	c.Close()
	c.Close() // idempotent
}
