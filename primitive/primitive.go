package primitive

import (
	"sync"
)

// LeftRightPrimitive provides the basic core of the leftt-right pattern.
type LeftRightPrimitive struct {
	// readIndicator is an array of 2 read-indicators, counting the reader numbers on the left/right instance
	lock *sync.RWMutex
	// readHere represents which instance to read
	read chan struct{}
	// other is the other instance
	other *LeftRightPrimitive
	Data  interface{}
}

// New creates a LeftRightPrimitive
func New(dataInit func() interface{}) *LeftRightPrimitive {
	r := &LeftRightPrimitive{
		lock: new(sync.RWMutex),
		read: make(chan struct{}, 1),
		Data: dataInit(),
	}

	l := &LeftRightPrimitive{
		lock:  new(sync.RWMutex),
		read:  make(chan struct{}, 1),
		Data:  dataInit(),
		other: r,
	}

	r.other = l
	l.read <- struct{}{}
	return l // starts reading on the left side
}

// ApplyReadFn applies read operation on the chosen instance, oh, I really need generics, interface type is ugly
func (lr *LeftRightPrimitive) ApplyReadFn(fn func(interface{})) {
	select {
	case m := <-lr.read:
		lr.read <- m
		lr.lock.RLock()
		fn(lr.Data)
		lr.lock.RUnlock()
	default:
		lr.other.ApplyReadFn(fn)
	}
}

// ApplyWriteFn applies write operation on the chosen instance, write operation is done twice, on the left and right
// instance respectively, this might make writing longer, but the readers are wait-free.
func (lr *LeftRightPrimitive) ApplyWriteFn(fn func(interface{})) {
	select {
	case m := <-lr.read:
		lr.read <- m
		lr.other.ApplyWriteFn(fn)
	default:
		lr.write(fn)
		m := <-lr.other.read
		lr.read <- m
		lr.other.write(fn)
	}
}

func (lr *LeftRightPrimitive) write(fn func(interface{})) {
	lr.lock.Lock()
	fn(lr.Data)
	lr.lock.Unlock()
}
