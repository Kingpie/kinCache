package lru

import "container/list"

type Cache struct {
	maxBytes    int64                         //最大字节
	curBytes    int64                         //当前已使用字节
	list        *list.List                    //队列
	cache       map[string]*list.Element      //缓存map
	delCallback func(key string, value Value) //删除回调
}

type Value interface {
	Len() int
}

type node struct {
	key   string
	value Value
}

// New creates a new Cache instance with the specified maximum size and delete callback.
// max: the maximum size of the cache in bytes
// callback: a function that will be called when an item is evicted from the cache,
//
//	receiving the key and value of the evicted item
//
// returns: a pointer to the newly created Cache instance
func New(max int64, callback func(string, Value)) *Cache {
	return &Cache{
		maxBytes:    max,
		list:        list.New(),
		cache:       make(map[string]*list.Element),
		delCallback: callback,
	}
}

// DelOldest 删除缓存中最旧的元素
// 该函数会从双向链表的尾部删除最久未访问的节点，
// 并从map中同步删除对应的键值对，同时更新当前缓存大小
func (c *Cache) DelOldest() {
	// 获取链表最后一个节点（最旧的元素）
	last := c.list.Back()
	if last != nil {
		// 获取节点存储的键值对
		kv := last.Value.(*node)
		// 从map中删除对应的键
		delete(c.cache, kv.key)
		// 从链表中移除该节点
		c.list.Remove(last)

		// 更新当前缓存大小，减去键和值占用的字节数
		c.curBytes -= int64(len(kv.key)) + int64(kv.value.Len())

		// 执行删除回调函数，通知调用方有元素被删除
		if c.delCallback != nil {
			c.delCallback(kv.key, kv.value)
		}
	}
}

// Get 从缓存中获取指定键的值
// 参数:	key - 要查找的键
//
// 返回值:
//
//	Value - 找到的值，如果未找到则为nil
//	bool - 表示是否找到该键，true表示找到，false表示未找到
func (c *Cache) Get(key string) (Value, bool) {
	if val, ok := c.cache[key]; ok {
		// 热数据移到队首
		c.list.MoveToFront(val)
		kv := val.Value.(*node)
		return kv.value, true
	}

	return nil, false
}

// Add 将指定的键值对添加到缓存中
// 如果键已存在，则更新其值并将其移到队首
// 如果键不存在，则添加新的键值对，如果缓存已满则删除最旧的元素
// key: 要添加的键
// value: 要添加的值
func (c *Cache) Add(key string, value Value) {
	//缓存击中移到队首
	if val, ok := c.cache[key]; ok {
		c.list.MoveToFront(val)

		//可能value已经改变
		kv := val.Value.(*node)
		c.curBytes += int64(kv.value.Len()) - int64(value.Len())
		kv.value = value
	} else {
		// 添加新的键值对到缓存
		item := c.list.PushFront(&node{key, value})
		c.cache[key] = item
		c.curBytes += int64(len(key)) + int64(value.Len())

		// 检查缓存是否超出最大容量，如果超出则删除最旧的元素
		for c.maxBytes < c.curBytes {
			c.DelOldest()
		}
	}
}
