package consistent_hash

import (
	"fmt"
	"testing"
)

func TestMap_AddNode(t *testing.T) {
	m := New(5, nil)
	m.AddNode("1", "2", "3")

	fmt.Printf("%+v", m.keys)

	testList := []string{"gege", "dada", "gaga", "mimi"}

	for _, s := range testList {
		t.Logf("%s local node:%s", s, m.Get(s))
	}
}
