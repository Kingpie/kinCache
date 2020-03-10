package consistent_hash

import (
	"fmt"
	"testing"
)

func TestMap_AddNode(t *testing.T) {
	m := New(5, nil)
	m.AddNode("hehe", "haha", "lala")

	fmt.Printf("%+v", m.keys)

	testList := []string{"xixi", "dada", "gaga", "mimi"}

	for _, s := range testList {
		t.Logf("%s local node:%s", s, m.Get(s))
	}
}
