package lru

import (
	"fmt"
	"testing"
)

type Hehe struct {
	A int
	B int
}

func (h Hehe) Len() int {
	return 8
}

func TestCache_Add(t *testing.T) {
	c := New(30, func(s string, value Value) {
		fmt.Printf("remove node,key:%s,value:%v\n", s, value)
	})
	c.Add("1", &Hehe{1, 1})
	c.Add("2", &Hehe{2, 2})
	c.Add("3", &Hehe{3, 3})
	c.Get("1") //查询之后变成热key
	c.Add("4", &Hehe{4, 4})
}
