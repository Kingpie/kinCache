package cache

import (
	"fmt"
	"github.com/Kingpie/kinCache/single_flight"
	"log"
	"sync"
)

type Group struct {
	name      string //组名
	getter    Getter //回调
	mainCache cache  //缓存
	peers     PeerPicker
	loader    *single_flight.Group //每个key同时只处理一个
}

var (
	rw     sync.RWMutex
	groups = make(map[string]*Group)
)

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("Register called more than once")
	}

	g.peers = peers
}

func (g *Group) load(key string) (value CacheValue, err error) {
	///每个key同时只处理一个请求
	val, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
			}
			log.Println("Failed to get from peer", err)
		}

		return g.getFromSrc(key)
	})

	if err == nil {
		return val.(CacheValue), nil
	}

	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (CacheValue, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return CacheValue{}, err
	}

	return CacheValue{b: bytes}, nil
}

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
		loader:    &single_flight.Group{},
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
