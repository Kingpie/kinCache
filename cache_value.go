package kinCache

//记录cache的内容
type CacheValue struct {
	b []byte
}

func (c CacheValue) Len() int {
	return len(c.b)
}

//复制一份value
func (c CacheValue) Copy() []byte {
	return clone(c.b)
}

func clone(b []byte) []byte {
	newVal := make([]byte, len(b))
	copy(newVal, b)
	return newVal
}

func (c CacheValue) String() string {
	return string(c.b)
}
