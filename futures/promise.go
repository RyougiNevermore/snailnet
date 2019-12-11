package futures

import (
	"fmt"
	"sync"
	"unsafe"
)

type Handler interface {
	Handle(ar AsyncResult)
}

type Promise interface {
	Handler
	Complete(p unsafe.Pointer)
	CompleteEmpty()
	Fail(err error)
}

var promisePool = &sync.Pool{New: func() interface{} {
	return NewPromise()
}}

func acquirePromise(f *future) Promise {
	p, _ := promisePool.Get().(*promise0)
	p.future = f
	return p
}

func releasePromise(p Promise) {
	p0 := (p).(*promise0)
	p0.err = nil
	p0.succeeded = false
	p0.failed = false
	p0.result = nil
	p0.future = nil
	promisePool.Put(p0)
}

func NewPromise() Promise {
	return &promise0{
		succeeded: false,
		failed:    false,
		result:    nil,
		err:       nil,
	}
}

type promise0 struct {
	succeeded bool
	failed    bool
	result    unsafe.Pointer
	err       error
	future    *future
}

func (p *promise0) Handle(ar AsyncResult) {
	if ar.Succeeded() {
		p.Complete(ar.Result())
	} else {
		p.Fail(ar.Error())
	}
}

func (p *promise0) Complete(v unsafe.Pointer) {
	defer releasePromise(p)

	if v == nil {
		panic(fmt.Errorf("complete promise failed, result is nil"))
	}

	if p.succeeded || p.failed {
		panic(fmt.Errorf("result is already complete: %s", func(succeeded bool) string {
			if succeeded {
				return "succeeded"
			}
			return "failed"
		}(p.succeeded)))
	}
	p.result = v
	p.succeeded = true

	ar := acquireAsyncResult(v, nil)

	p.future.send(ar)

}

func (p *promise0) CompleteEmpty() {
	p.Complete(_voidPtr)
}

func (p *promise0) Fail(err error) {
	defer releasePromise(p)
	if err == nil {
		panic(fmt.Errorf("faile promise failed, error is nil"))
	}
	if p.succeeded || p.failed {
		panic(fmt.Errorf("result is already complete: %s", func(succeeded bool) string {
			if succeeded {
				return "succeeded"
			}
			return "failed"
		}(p.succeeded)))
	}
	p.err = err
	p.succeeded = true

	ar := acquireAsyncResult(nil, err)

	p.future.send(ar)
}
