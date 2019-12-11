package futures

import (
	"fmt"
	"sync"
)

var futurePool = &sync.Pool{New: func() interface{} {
	return &future{
		ch:      make(chan *asyncResult0, 1),
		handler: nil,
	}
}}

func acquireFuture(h Handler) *future {
	f, _ := futurePool.Get().(*future)
	f.handler = h
	return f
}

func releaseFuture(f *future) {
	f.handler = nil
	futurePool.Put(f)
}

type future struct {
	ch      chan *asyncResult0
	handler Handler
}

func (f *future) send(ar *asyncResult0) {
	f.ch <- ar
}

func (f *future) wait() {
	p, ok := <-f.ch
	if !ok {
		panic(fmt.Errorf("future wait result failed, buffer is shutdown"))
	}
	f.handler.Handle(p)
	releaseAsyncResult(p)
}
