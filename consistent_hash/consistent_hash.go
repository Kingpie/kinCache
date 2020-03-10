package consistent_hash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

//哈希算法
type Hash func(data []byte) uint32

type Map struct {
	hashFunc Hash
	replicas int
	keys     []int
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hashFunc: fn,
		hashMap:  make(map[int]string),
	}

	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

//增加机器节点,生成replicas个虚拟节点
func (m *Map) AddNode(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hashFunc([]byte(key + strconv.Itoa(i))))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

//寻找最近的节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hashFunc([]byte(key)))

	//二分查找
	index := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	//环状所以取余
	return m.hashMap[m.keys[index%len(m.keys)]]
}
