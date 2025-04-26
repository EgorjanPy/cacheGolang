package main

import (
	"fmt"
	"sync"
	"time"
)

type CacheEntry struct {
	value    interface{}
	expireAt int64
}

func NewCacheEntry(value interface{}, expireAt int64) CacheEntry {
	return CacheEntry{
		value:    value,
		expireAt: expireAt,
	}
}

func (ce CacheEntry) IsExpired() bool {
	return ce.expireAt < time.Now().UnixNano()
}

type Cache struct {
	kvstore  map[string]CacheEntry
	locker   sync.RWMutex
	interval time.Duration
	stop     chan struct{}
}

func NewCache(cleanUpInterval time.Duration) *Cache {
	cache := &Cache{
		kvstore:  make(map[string]CacheEntry),
		interval: cleanUpInterval,
		stop:     make(chan struct{}),
	}

	if cleanUpInterval > 0 {
		go cache.cleaning()
	}
	return cache
}

func (c *Cache) cleaning() {
	fmt.Println("cleaner starting...")
	ticker := time.NewTicker(c.interval)
	fmt.Println("cleaner was started")
	for {
		select {
		case <-ticker.C:
			c.purge()
		case <-c.stop:
			ticker.Stop()
			fmt.Println("cleaner was stopped")
			return
		}
	}
}

func (c *Cache) purge() {
	c.locker.Lock()
	defer c.locker.Unlock()
	for key, data := range c.kvstore {
		if data.IsExpired() {
			delete(c.kvstore, key)
		}
	}
}

func (c *Cache) Set(key string, value interface{}, expiryDuration time.Duration) {
	expireAt := time.Now().Add(expiryDuration).UnixNano()
	c.locker.Lock()
	defer c.locker.Unlock()
	c.kvstore[key] = NewCacheEntry(value, expireAt)

}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.locker.RLock()
	defer c.locker.RUnlock()
	data, found := c.kvstore[key]
	if !found || data.IsExpired() {
		return nil, false
	}

	return data.value, true
}

func (c *Cache) Close() {
	close(c.stop)
}

func main() {
	cache := NewCache(time.Second)
	defer cache.Close()
	cache.Set("foo", "bar", 2*time.Second)
	for i := 0; i < 3; i++ {
		value, found := cache.Get("foo")
		if found {
			fmt.Println("value for key foo is ", value)
		} else {
			fmt.Println("value for key foo is not found")
			break
		}

		fmt.Println("waiting for 1 second...")
		time.Sleep(time.Second)
	}
}
