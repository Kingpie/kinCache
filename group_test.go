package kinCache

import (
	"fmt"
	"testing"
)

var mydb = map[string]string{
	"a": "1",
	"b": "1",
	"c": "1",
	"d": "1",
}

func TestGroup_Get(t *testing.T) {
	g := NewGroup("test", 10, GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := mydb[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	data, _ := g.Get("a")
	_, _ = g.Get("a")
	t.Logf("%s", data.String())
}
