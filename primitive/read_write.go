package primitive

type Reader interface {
	Read(interface{})
}
type ReaderFunc func(interface{})

func (r ReaderFunc) Read(i interface{}) {
	r(i)
}

type Writer interface {
	Write(interface{})
}
type WriterFunc func(interface{})

func (w WriterFunc) Write(i interface{}) {
	w(i)
}
