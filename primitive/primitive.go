package primitive

import (
	"sync"
)

type LeftRight struct {
	*sync.RWMutex
	read  *leftRightPrimitive
	write *leftRightPrimitive
}

// leftRightPrimitive provides the basic core of the left-right pattern.
type leftRightPrimitive struct {
	// lock protects this side
	*sync.RWMutex
	data interface{}
}

func (p *leftRightPrimitive) read(reader Reader) {
	p.RLock()
	reader.Read(p.data)
	p.RUnlock()
}

func (p *leftRightPrimitive) write(writer Writer) {
	p.Lock()
	writer.Write(p.data)
	p.Unlock()
}

// New creates a LeftRightPrimitive
func New(dataInit func() interface{}) *LeftRight {
	r := &leftRightPrimitive{
		new(sync.RWMutex),
		dataInit(),
	}

	l := &leftRightPrimitive{
		new(sync.RWMutex),
		dataInit(),
	}

	lr := &LeftRight{
		RWMutex: new(sync.RWMutex),
		read:    l,
		write:   r,
	}
	return lr
}

// ApplyReadFn applies read operation on the chosen instance, oh, I really need generics, interface type is ugly
func (lr *LeftRight) ApplyReadFn(fn func(interface{})) {
	lr.ApplyReader(ReaderFunc(fn))
}

// ApplyReader applies read operation on the chosen instance, oh, I really need generics, interface type is ugly
func (lr *LeftRight) ApplyReader(fn Reader) {
	lr.getReader().read(fn)
}

// ApplyWriteFn applies write operation on the chosen instance, write operation is done twice, on the left and right
// instance respectively, this might make writing longer, but the readers are wait-free.
func (lr *LeftRight) ApplyWriteFn(fn func(interface{})) {
	lr.ApplyWriter(WriterFunc(fn))
}

// ApplyWriter applies write operation on the chosen instance, write operation is done twice, on the left and right
// instance respectively, this might make writing longer, but the readers are wait-free.
func (lr *LeftRight) ApplyWriter(fn Writer) {
	reader, writer := lr.getReader(), lr.getWriter()
	writer.write(fn)
	go reader.write(fn)
	lr.swap()
}

func (lr *LeftRight) swap() {
	lr.Lock()
	defer lr.Unlock()
	lr.read, lr.write = lr.write, lr.read
}

func (lr *LeftRight) getReader() *leftRightPrimitive {
	lr.RLock()
	defer lr.RUnlock()
	return lr.read
}

func (lr *LeftRight) getWriter() *leftRightPrimitive {
	lr.RLock()
	defer lr.RUnlock()
	return lr.write
}

func (lr *LeftRight) getData() interface{} {
	return lr.getReader().data
}
