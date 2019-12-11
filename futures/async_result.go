package futures

import (
	"sync"
	"unsafe"
)

type AsyncResult interface {
	Result() (p unsafe.Pointer)
	Error() (err error)
	Succeeded() bool
	Failed() bool
	Map(fn func(p unsafe.Pointer) unsafe.Pointer) (ar AsyncResult)
	Otherwise(fn func(err error) unsafe.Pointer) (ar AsyncResult)
}

var asyncResultPool = &sync.Pool{New: func() interface{} {
	return &asyncResult0{}
}}

type asyncResult0 struct {
	p   unsafe.Pointer
	err error
}

func (ar *asyncResult0) Result() unsafe.Pointer {
	return ar.p
}

func (ar *asyncResult0) Error() error {
	return ar.err
}

func (ar *asyncResult0) Succeeded() bool {
	return ar.err == nil
}

func (ar *asyncResult0) Failed() bool {
	return ar.err != nil
}

func (ar *asyncResult0) Map(fn func(p unsafe.Pointer) unsafe.Pointer) (ar0 *asyncResult0) {
	ar0, _ = asyncResultPool.Get().(*asyncResult0)
	if ar.err != nil {
		return
	}
	ar0.p = fn(ar.p)
	return
}

func (ar *asyncResult0) Otherwise(fn func(err error) unsafe.Pointer) (ar0 *asyncResult0) {
	if ar.err == nil {
		ar0 = ar
		return
	}
	ar0, _ = asyncResultPool.Get().(*asyncResult0)
	ar0.p = fn(ar.err)
	return
}
