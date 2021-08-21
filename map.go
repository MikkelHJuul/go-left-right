package lrc

import "github.com/csimplestring/go-left-right/primitive"

// LRMap utilises the left-right pattern to handle concurrent read-write.
type LRMap struct {
	*primitive.LeftRightPrimitive

	left  map[int]int
	right map[int]int
}

func newIntMap() *LRMap {

	return &LRMap{LeftRightPrimitive: primitive.New(func() interface{} {
		return make(map[int]int)
	})}
}

func (lr *LRMap) Get(k int) (val int, exist bool) {

	lr.ApplyReadFn(func(instance interface{}) {
		m := instance.(map[int]int)
		val, exist = m[k]
	})

	return
}

func (lr *LRMap) Put(key, val int) {
	lr.ApplyWriteFn(func(instance interface{}) {
		m := instance.(map[int]int)
		m[key] = val
	})
}
