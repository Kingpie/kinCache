package cache

import "testing"

func TestGetterFunc_Get(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	v, _ := f.Get("key")
	t.Logf("key=%s", v)
}
