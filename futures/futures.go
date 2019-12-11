package futures

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

func Init() {
	group = New(0)
}

var group *Group

func Future(handler Handler) Promise {
	if group == nil {
		panic(fmt.Errorf("make a future failed, use Init() at first, and use CloseAwait() at last"))
	}
	return group.Future(handler)
}

func CloseAwait(ctx context.Context) {
	group.CloseAwait(ctx)
}

func New(workers int) *Group {
	if workers <= 0 {
		workers = runtime.NumCPU() * 2
	}
	length := int64(1024 * workers)
	buffers := make([]*arrayBuffer, 0, 1)
	for i := 0; i < workers; i++ {
		buffer := newArrayBuffer(length)
		buffers = append(buffers, buffer)
		go func(buffer *arrayBuffer) {
			for {
				v, ok := buffer.Recv()
				if !ok {
					break
				}
				f := (*future)(v)
				f.wait()
				releaseFuture(f)
			}
		}(buffer)
	}
	return &Group{
		running:   1,
		runningMx: sync.Mutex{},
		idx:       0,
		mask:      uint64(workers),
		buffers:   buffers,
	}
}

type Group struct {
	running   int64
	runningMx sync.Mutex
	idx       uint64
	mask      uint64
	buffers   []*arrayBuffer
}

func (g *Group) Future(handler Handler) Promise {
	if atomic.LoadInt64(&g.running) == 0 {
		panic(fmt.Errorf("group is not running"))
		return nil
	}
	if handler == nil {
		panic(fmt.Errorf("handler is nil"))
	}
	f := acquireFuture(handler)

	g.next(f)

	p := acquirePromise(f)
	return p
}

func (g *Group) next(f *future) {
	atomic.AddUint64(&g.idx, 1)
	x := g.idx & g.mask
	buffer := g.buffers[x]
	if err := buffer.Send(unsafe.Pointer(f)); err != nil {
		panic(fmt.Errorf("make a future failed, %w", err))
	}
}

func (g *Group) CloseAwait(ctx context.Context) {
	if !atomic.CompareAndSwapInt64(&g.running, 1, 0) {
		return
	}
	for _, buffer := range g.buffers {
		_ = buffer.Close()
	}
	for _, buffer := range g.buffers {
		_ = buffer.Sync(ctx)
	}
}
