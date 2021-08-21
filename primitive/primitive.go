package primitive

import (
	"sync"
)

// LeftRightPrimitive provides the basic core of the leftt-right pattern.
type LeftRightPrimitive struct {
	// readIndicator is an array of 2 read-indicators, counting the reader numbers on the left/right instance
	lock *sync.RWMutex
	// readHere represents which instance to read
	writeJobs chan func(interface{})
	// other is the other instance
	other *LeftRightPrimitive
	Data  interface{}
}

// New creates a LeftRightPrimitive
func New(leftData interface{}, rightData interface{}) *LeftRightPrimitive {
        writeJobs := make(chan func(interface{}))
	r := &LeftRightPrimitive{
		lock:          sync.RWMutex{},
		writeJobs:     writeJobs,
		Data:          rightData,
	}

	l := &LeftRightPrimitive{
		lock:          sync.RWMutex{},
		writeJobs:     writeJobs,
		Data:          leftData,
		other:         r,
	}

	r.other = l
	return l // starts reading on the left side
}

// ApplyReadFn applies read operation on the chosen instance, oh, I really need generics, interface type is ugly
func (lr *LeftRightPrimitive) ApplyReadFn(fn func(interface{})) {
	select {
	case wrt := <- lr.writeJobs:
		lr.other.ApplyReadFn(fn)
		lr.write(wrt)
	default:
		lr.lock.RLock()
		fn(lr.Data)
		lr.lock.RUlock()
	}
}

// ApplyWriteFn applies write operation on the chosen instance, write operation is done twice, on the left and right
// instance respectively, this might make writing longer, but the readers are wait-free.
func (lr *LeftRightPrimitive) ApplyWriteFn(fn func(interface{})) {
	select {
	case wrt := <- writeJobs:
		lr.write(wrt)
		lr.write(fn)
		writeJobs <- fn
	default:
                lr.other.write(fn)
        }
}

func (lr *LeftRightPrimitive) write(fn func(interface{})) {
        lr.lock.Lock()
        fn(lr.Data)
        lr.lock.Unlock()
}
