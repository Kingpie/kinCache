package cache

import (
	"fmt"
	"log"
	"sync"
)

type Group struct {
	name      string //组名
	getter    Getter //回调
	mainCache cache  //缓存
}

var (
	rw     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, bytes int64, getter Getter) *Group {
	if getter == nil {
		panic("getter is nil")
	}

	rw.Lock()
	defer rw.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: bytes},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	rw.RLock()
	defer rw.RLock()

	g := groups[name]
	return g
}

func (g *Group) Get(key string) (CacheValue, error) {
	if key == "" {
		return CacheValue{}, fmt.Errorf("key is empty")
	}

	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (CacheValue, error) {
	return g.getFromSrc(key)
}

//从源获取数据
func (g *Group) getFromSrc(key string) (CacheValue, error) {
	data, err := g.getter.Get(key)
	if err != nil {
		return CacheValue{}, err
	}

	log.Printf("getFromSrc key:%s", key)

	value := CacheValue{b: clone(data)}
	g.insertCache(key, value)
	return value, nil
}

func (g *Group) insertCache(key string, value CacheValue) {
	g.mainCache.add(key, value)
}
