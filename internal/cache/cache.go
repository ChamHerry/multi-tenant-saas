package cache

import (
	"sync"
	"time"
)

// Item holds a cached value with an expiration time.
type Item[T any] struct {
	Value T
	Exp   time.Time
}

// Cache is a simple in-memory TTL cache safe for concurrent use.
type Cache[T any] struct {
	mu    sync.RWMutex
	items map[string]Item[T]
}

// New creates a new Cache instance.
func New[T any]() *Cache[T] {
	return &Cache[T]{items: map[string]Item[T]{}}
}

// Get returns the cached value or the zero value if missing or expired.
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	if !ok || time.Now().After(item.Exp) {
		var zero T
		return zero, false
	}
	return item.Value, true
}

// Set stores a value with the given TTL.
func (c *Cache[T]) Set(key string, value T, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Item[T]{Value: value, Exp: time.Now().Add(ttl)}
}

// Invalidate removes a specific key.
func (c *Cache[T]) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// InvalidatePrefix removes all keys with the given prefix.
func (c *Cache[T]) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.items, k)
		}
	}
}
