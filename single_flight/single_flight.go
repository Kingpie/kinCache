package single_flight

import "sync"

//正在进行或者已结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

//管理不同key的请求
type Group struct {
	mtx sync.Mutex
	m   map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mtx.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mtx.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mtx.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mtx.Lock()
	delete(g.m, key)
	g.mtx.Unlock()

	return c.val, c.err
}
