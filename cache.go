package main

import (
	"log"
	"sync"
	"time"
)

type CacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

type Cache struct {
	sync.RWMutex
	items map[string]CacheEntry
}

func NewCache(cleanupFrequency time.Duration) *Cache {
	cache := &Cache{
		items: make(map[string]CacheEntry),
	}
	// Start a background goroutine to clean expired items
	go cache.cleanupLoop(cleanupFrequency)
	return cache
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.Lock()
	defer c.Unlock()

	c.items[key] = CacheEntry{
		data:      value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		return nil, false
	}

	return item.data, true
}

func (c *Cache) cleanupLoop(cleanupFrequency time.Duration) {
	ticker := time.NewTicker(cleanupFrequency)
	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache) cleanup() {
	c.Lock()
	defer c.Unlock()

	log.Println("Running cleanup")
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
}
