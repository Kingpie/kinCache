package lru

import "container/list"

type Cache struct {
	maxBytes int64                         //最大字节
	curBytes int64                         //当前已使用字节
	list     *list.List                    //队列
	cache    map[string]*list.Element      //缓存map
	callback func(key string, value Value) //删除回调
}

type Value interface {
	Len() int
}

type node struct {
	key   string
	value Value
}

func New(max int64, callback func(string, Value)) *Cache {
	return &Cache{
		maxBytes: max,
		list:     list.New(),
		cache:    make(map[string]*list.Element),
		callback: callback,
	}
}

//删除最后的元素
func (c *Cache) DelOldest() {
	last := c.list.Back()
	if last != nil {
		kv := last.Value.(*node)
		delete(c.cache, kv.key)
		c.list.Remove(last)

		c.curBytes -= int64(len(kv.key)) + int64(kv.value.Len())

		//执行回调
		if c.callback != nil {
			c.callback(kv.key, kv.value)
		}
	}
}

//查询元素
func (c *Cache) Get(key string) (Value, bool) {
	if val, ok := c.cache[key]; ok {
		//热数据移到队首
		c.list.MoveToFront(val)
		kv := val.Value.(*node)
		return kv.value, true
	}

	return nil, false
}

//增加元素
func (c *Cache) Add(key string, value Value) {
	//缓存击中移到队首
	if val, ok := c.cache[key]; ok {
		c.list.MoveToFront(val)

		//可能value已经改变
		kv := val.Value.(*node)
		c.curBytes += int64(kv.value.Len()) - int64(value.Len())
		kv.value = value
	} else {
		item := c.list.PushFront(&node{key, value})
		c.cache[key] = item
		c.curBytes += int64(len(key)) + int64(value.Len())

		for c.maxBytes < c.curBytes {
			c.DelOldest()
		}
	}
}
