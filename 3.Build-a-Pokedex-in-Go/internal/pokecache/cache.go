package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu      sync.Mutex
	pokemap map[string]cacheEntry
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		for key, val := range c.pokemap {
			if time.Since(val.createdAt) >= interval {
				delete(c.pokemap, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pokemap[key] = cacheEntry{
		val:       val,
		createdAt: time.Now(),
	}

}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	pokemap, ok := c.pokemap[key]
	if !ok {
		return nil, false
	}

	return pokemap.val, true

}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		mu:      sync.Mutex{},
		pokemap: make(map[string]cacheEntry),
	}
	go cache.reapLoop(interval)
	return cache
}
