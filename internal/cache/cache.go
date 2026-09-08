package cache

import (
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiration int64
}

// MemoryCache provides a fast, thread-safe in-memory cache with per-item TTL
type MemoryCache struct {
	items sync.Map
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{}
	// Run background cleaner every 3 minutes
	go c.startJanitor(3 * time.Minute)
	return c
}

// Get retrieves an item by key. Returns value and true if found and not expired.
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	val, ok := c.items.Load(key)
	if !ok {
		return nil, false
	}

	it := val.(item)
	if it.expiration > 0 && time.Now().UnixNano() > it.expiration {
		c.items.Delete(key)
		return nil, false
	}

	return it.value, true
}

// Set saves an item with a specified TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.items.Store(key, item{
		value:      value,
		expiration: exp,
	})
}

// Delete removes an item
func (c *MemoryCache) Delete(key string) {
	c.items.Delete(key)
}

// Clear clears all cache
func (c *MemoryCache) Clear() {
	c.items.Range(func(key, value interface{}) bool {
		c.items.Delete(key)
		return true
	})
}

func (c *MemoryCache) startJanitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		now := time.Now().UnixNano()
		c.items.Range(func(key, value interface{}) bool {
			it := value.(item)
			if it.expiration > 0 && now > it.expiration {
				c.items.Delete(key)
			}
			return true
		})
	}
}
