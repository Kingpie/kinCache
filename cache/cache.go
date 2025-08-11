package cache

import (
	"kinCache/lru"
	"sync"
)

type cache struct {
	mtx        sync.Mutex
	lru        *lru.Cache
	cacheBytes int64 //缓存大小
}

func (c *cache) add(key string, value CacheValue) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (val CacheValue, ok bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(CacheValue), ok
	}
	return
}
