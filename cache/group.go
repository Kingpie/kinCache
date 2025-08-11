package cache

import (
	"fmt"
	"kinCache/single_flight"
	"log"
	"sync"
)

type Group struct {
	name      string               //组名
	getter    Getter               //回调
	mainCache cache                //缓存
	peers     PeerPicker           //集群节点
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

// load 从缓存中加载指定key的值，如果本地没有则从其他节点或数据源获取
// 参数:
//
//	key: 需要获取的缓存键
//
// 返回值:
//
//	value: 获取到的缓存值
//	err: 获取过程中发生的错误
func (g *Group) load(key string) (value CacheValue, err error) {
	// 每个key同时只处理一个请求，避免缓存击穿
	val, err := g.loader.Do(key, func() (interface{}, error) {
		// 如果配置了集群节点，则优先从其他节点获取数据
		if g.peers != nil {
			// 根据key选择合适的节点
			if peer, ok := g.peers.PickPeer(key); ok {
				// 从选中的节点获取数据
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
			}
			log.Println("Failed to get from peer", err)
		}

		// 从数据源获取数据
		return g.getFromSrc(key)
	})

	// 类型转换并返回结果
	if err == nil {
		return val.(CacheValue), nil
	}

	return
}

// getFromPeer 从指定的PeerGetter获取缓存数据
// 参数:
//
//	peer: PeerGetter接口，用于从远程节点获取数据
//	key: string类型，要获取的缓存键
//
// 返回值:
//
//	CacheValue: 获取到的缓存值
//	error: 获取过程中发生的错误
func (g *Group) getFromPeer(peer PeerGetter, key string) (CacheValue, error) {
	// 调用peer的Get方法获取数据
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return CacheValue{}, err
	}

	// 将获取到的字节数据封装为CacheValue并返回
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

	//先查询本地缓存
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}

	// 缓存中没有，则从源获取数据
	return g.load(key)
}

// 从源获取数据
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
