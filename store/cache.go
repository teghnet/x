package store

import (
	"sync"
	"time"
)

// Cache is a concurrency-safe, in-memory key/value cache with per-entry
// time-to-live. Expired entries are removed lazily on access and, if a janitor
// is started, periodically in the background. The zero value is not usable;
// call NewCache.
type Cache[K comparable, V any] struct {
	mu      sync.Mutex
	items   map[K]entry[V]
	now     func() time.Time // injectable clock for testing
	stop    chan struct{}
	stopped bool
}

type entry[V any] struct {
	value   V
	expires time.Time // zero means no expiry
}

// NewCache returns an empty cache.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]entry[V]),
		now:   time.Now,
	}
}

// Set stores value under key with the given ttl. A ttl <= 0 means the entry
// never expires.
func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := entry[V]{value: value}
	if ttl > 0 {
		e.expires = c.now().Add(ttl)
	}
	c.items[key] = e
}

// Get returns the value for key and whether it was present and unexpired.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[key]
	if !ok || c.expired(e) {
		if ok {
			delete(c.items, key)
		}
		var zero V
		return zero, false
	}
	return e.value, true
}

// GetOrLoad returns the cached value for key. On a miss it calls load, stores
// the result with ttl, and returns it. load runs while the cache lock is held,
// so it must not call back into the cache; keep it short.
func (c *Cache[K, V]) GetOrLoad(key K, ttl time.Duration, load func() (V, error)) (V, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.items[key]; ok && !c.expired(e) {
		return e.value, nil
	}
	v, err := load()
	if err != nil {
		return v, err
	}
	e := entry[V]{value: v}
	if ttl > 0 {
		e.expires = c.now().Add(ttl)
	}
	c.items[key] = e
	return v, nil
}

// Delete removes key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Len returns the number of unexpired entries, purging expired ones.
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.purgeLocked()
	return len(c.items)
}

func (c *Cache[K, V]) expired(e entry[V]) bool {
	return !e.expires.IsZero() && !c.now().Before(e.expires)
}

func (c *Cache[K, V]) purgeLocked() {
	for k, e := range c.items {
		if c.expired(e) {
			delete(c.items, k)
		}
	}
}

// StartJanitor launches a goroutine that purges expired entries every interval
// until Close is called. It is optional: without it, expiry is still enforced
// lazily on access.
func (c *Cache[K, V]) StartJanitor(interval time.Duration) {
	c.mu.Lock()
	if c.stop != nil || c.stopped {
		c.mu.Unlock()
		return
	}
	c.stop = make(chan struct{})
	stop := c.stop
	c.mu.Unlock()

	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				c.mu.Lock()
				c.purgeLocked()
				c.mu.Unlock()
			}
		}
	}()
}

// Close stops the janitor goroutine, if any. It is safe to call more than once.
func (c *Cache[K, V]) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopped {
		return
	}
	c.stopped = true
	if c.stop != nil {
		close(c.stop)
		c.stop = nil
	}
}
