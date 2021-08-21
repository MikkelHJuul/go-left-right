package primitive

import (
	"sync"
	"sync/atomic"
)

const ReadHere int32 = 1
const ReadOther int32 = 0

// LeftRightPrimitive provides the basic core of the leftt-right pattern.
type LeftRightPrimitive struct {
	// lock protects this side
	lock *sync.RWMutex
	// readHere represents which instance to read
	readHere *int32
	// other is the other instance
	other *LeftRightPrimitive
	Data  interface{}
}

// New creates a LeftRightPrimitive
func New(leftData interface{}, rightData interface{}) *LeftRightPrimitive {
	r := &LeftRightPrimitive{
		lock:     new(sync.RWMutex),
		readHere: new(int32),
		Data:     rightData,
	}

	l := &LeftRightPrimitive{
		lock:     new(sync.RWMutex),
		readHere: new(int32),
		Data:     leftData,
		other:    r,
	}

	*l.readHere = ReadHere
	*r.readHere = ReadOther
	r.other = l
	return l // starts reading on the left side
}

// ApplyReadFn applies read operation on the chosen instance, oh, I really need generics, interface type is ugly
func (lr *LeftRightPrimitive) ApplyReadFn(fn func(interface{})) {
	if lr.isReader() {
		lr.read(fn)
	} else {
		lr.other.read(fn)
	}
}

func (lr *LeftRightPrimitive) read(fn func(interface{})) {
	lr.lock.RLock()
	fn(lr.Data)
	lr.lock.RUnlock()
}

// ApplyWriteFn applies write operation on the chosen instance, write operation is done twice, on the left and right
// instance respectively, this might make writing longer, but the readers are wait-free.
func (lr *LeftRightPrimitive) ApplyWriteFn(fn func(interface{})) {
	if lr.isReader() {
		writeAndToggle(lr.other, lr, fn)
	} else {
		writeAndToggle(lr, lr.other, fn)
	}
}

func writeAndToggle(first, second *LeftRightPrimitive, fn func(interface{})) {
	first.write(fn)
	first.startRead()
	second.stopRead()
	second.write(fn)
}

func (lr *LeftRightPrimitive) write(fn func(interface{})) {
	lr.lock.Lock()
	fn(lr.Data)
	lr.lock.Unlock()
}

func (lr *LeftRightPrimitive) startRead() {
	atomic.StoreInt32(lr.readHere, ReadHere)
}

func (lr *LeftRightPrimitive) stopRead() {
	atomic.StoreInt32(lr.readHere, ReadOther)
}

func (lr *LeftRightPrimitive) isReader() bool {
	read := atomic.LoadInt32(lr.readHere)
	return read == ReadHere
}
