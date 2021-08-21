package primitive

import (
	"sync"
	"sync/atomic"
)

const ReadHere int32 = 1
const ReadOther int32 = 0

// LeftRightPrimitive provides the basic core of the leftt-right pattern.
type LeftRightPrimitive struct {
	// readIndicator is an array of 2 read-indicators, counting the reader numbers on the left/right instance
	lock *RWMutex
	// readHere represents which instance to read
	readHere *int32
	// other is the other instance
	other *LeftRightPrimitive
	Data  interface{}
}

// New creates a LeftRightPrimitive
func New(leftData interface{}, rightData interface{}) *LeftRightPrimitive {
	r := &LeftRightPrimitive{
		lock:          sync.RWMutex{},
		readHere:      new(int32),
		Data:          rightData,
	}

	l := &LeftRightPrimitive{
		lock:          sync.RWMutex{},
		readHere:      new(int32),
		Data:          leftData,
		other:         r,
	}

	*l.readHere = ReadHere
	*r.readHere = ReadOther
	r.other = l
	return l // starts reading on the left side
}

// ApplyReadFn applies read operation on the chosen instance, oh, I really need generics, interface type is ugly
func (lr *LeftRightPrimitive) ApplyReadFn(fn func(interface{})) {
	if lr.isReader() {
		lr.lock.RLock()
		fn(lr.Data)
                lr.lock.RUlock()
	} else {
                lr.other.ApplyReadFn(fn)
        }
}

// ApplyWriteFn applies write operation on the chosen instance, write operation is done twice, on the left and right
// instance respectively, this might make writing longer, but the readers are wait-free.
func (lr *LeftRightPrimitive) ApplyWriteFn(fn func(interface{})) {
	if lr.isReader() {
		lr.other.write(fn)
		lr.other.startRead()
                lr.stopRead()
		lr.write(fn)
	} else {
                lr.other.ApplyWriteFn(fn)
        }
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
