package futures

import (
	"sync"
	"unsafe"
)

type Void struct{}

var _void = &Void{}
var _voidPtr = unsafe.Pointer(_void)

type AsyncResult interface {
	Result() (p unsafe.Pointer)
	Error() (err error)
	Succeeded() bool
	Failed() bool
	IsEmpty() bool
	Map(fn func(p unsafe.Pointer) unsafe.Pointer) (ar AsyncResult)
	Otherwise(fn func(err error) unsafe.Pointer) (ar AsyncResult)
}

var asyncResultPool = &sync.Pool{New: func() interface{} {
	return &asyncResult0{}
}}

func acquireAsyncResult(p unsafe.Pointer, err error) *asyncResult0 {
	ar, _ := asyncResultPool.Get().(*asyncResult0)
	ar.p = p
	ar.err = err
	return ar
}

func releaseAsyncResult(ar *asyncResult0) {
	ar.p = nil
	ar.err = nil
	asyncResultPool.Put(ar)
}

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

func (ar *asyncResult0) IsEmpty() bool {
	return ar.p == _voidPtr
}

func (ar *asyncResult0) Map(fn func(p unsafe.Pointer) unsafe.Pointer) (ar0 AsyncResult) {
	ar0, _ = asyncResultPool.Get().(AsyncResult)
	if ar.err != nil {
		return
	}
	ar0.(*asyncResult0).p = fn(ar.p)
	return
}

func (ar *asyncResult0) Otherwise(fn func(err error) unsafe.Pointer) (ar0 AsyncResult) {
	if ar.err == nil {
		ar0 = ar
		return
	}
	ar0, _ = asyncResultPool.Get().(*asyncResult0)
	ar0.(*asyncResult0).p = fn(ar.err)
	return
}
