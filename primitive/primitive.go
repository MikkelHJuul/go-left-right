package primitive

import (
	"sync"
	"time"
)

type LeftRight struct {
	left   *leftRightPrimitive
	right  *leftRightPrimitive
	reads  chan<- Reader
	writes chan<- Writer
	done   chan struct{}
	cfg    config
}

func (lr *LeftRight) init(readChan <-chan Reader, writeChan <-chan Writer) {
	head, tail := lr.left, lr.right
	pendingWrites := make(chan Writer, lr.cfg.maxWrites)
	emptyPendingWrites := func(writeTo *leftRightPrimitive, pending <-chan Writer) {
		for {
			select {
			case fn := <-pending:
				go writeTo.write(fn)
			default:
				goto Return
			}
		}
	Return:
	}
	for {
		select {
		case readFun := <-readChan:
			go head.read(readFun)
		case writerFun := <-writeChan:
			go tail.write(writerFun)
			select {
			case pendingWrites <- writerFun:
			default:
				head, tail = tail, head
				go tail.write(writerFun)
				emptyPendingWrites(tail, pendingWrites)
			}
		case <-lr.cfg.ticker.C:
			if len(pendingWrites) != 0 {
				head, tail = tail, head
				emptyPendingWrites(tail, pendingWrites)
			}
		case <-lr.done:
			goto End
		}
	}
End:
}

// leftRightPrimitive provides the basic core of the left-right pattern.
type leftRightPrimitive struct {
	// lock protects this side
	*sync.RWMutex
	Data interface{}
}

func (p *leftRightPrimitive) read(reader Reader) {
	p.RLock()
	reader.Read(p.Data)
	p.RUnlock()
}

func (p *leftRightPrimitive) write(writer Writer) {
	p.Lock()
	writer.Write(p.Data)
	p.Unlock()
}

func NewDefault(dataInit func() interface{}) *LeftRight {
	return New(dataInit, WithMaxNumWritesPerSync(10), WithMaxSyncDuration(10*time.Microsecond))
}

// New creates a LeftRightPrimitive
func New(dataInit func() interface{}, confModifers ...func(*config)) *LeftRight {
	r := &leftRightPrimitive{
		new(sync.RWMutex),
		dataInit(),
	}

	l := &leftRightPrimitive{
		new(sync.RWMutex),
		dataInit(),
	}
	reads := make(chan Reader)
	writes := make(chan Writer)

	config := &config{}
	for _, m := range confModifers {
		m(config)
	}
	if config.ticker == nil {
		WithMaxSyncDuration(10 * time.Microsecond)(config)
	}

	lr := &LeftRight{
		left:   l,
		right:  r,
		reads:  reads,
		writes: writes,
		done:   make(chan struct{}),
		cfg:    *config,
	}

	go lr.init(reads, writes)
	return lr
}

// ApplyReadFn applies read operation on the chosen instance, oh, I really need generics, interface type is ugly
func (lr *LeftRight) ApplyReadFn(fn func(interface{})) {
	lr.ApplyReader(ReaderFunc(fn))
}

// ApplyReader applies read operation on the chosen instance, oh, I really need generics, interface type is ugly
func (lr *LeftRight) ApplyReader(fn Reader) {
	lr.reads <- fn
}

// ApplyWriteFn applies write operation on the chosen instance, write operation is done twice, on the left and right
// instance respectively, this might make writing longer, but the readers are wait-free.
func (lr *LeftRight) ApplyWriteFn(fn func(interface{})) {
	lr.ApplyWriter(WriterFunc(fn))
}

// ApplyWriter applies write operation on the chosen instance, write operation is done twice, on the left and right
// instance respectively, this might make writing longer, but the readers are wait-free.
func (lr *LeftRight) ApplyWriter(fn Writer) {
	lr.writes <- fn
}

func (lr *LeftRight) Close() error {
	close(lr.reads)
	close(lr.writes)
	return lr.cfg.Close()
}
