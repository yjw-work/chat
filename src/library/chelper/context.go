package chelper

import (
	"sync"
	"sync/atomic"
)

type TContext struct {
	WG     sync.WaitGroup
	chDone chan struct{}
	closed int32
}

func NewContext() *TContext {
	return &TContext{
		chDone: make(chan struct{}),
	}
}

func (this *TContext) Done() <-chan struct{} {
	return this.chDone
}

func (this *TContext) CloseDone() bool {
	if atomic.CompareAndSwapInt32(&this.closed, 0, 1) {
		close(this.chDone)
		return true
	}
	return false
}

func (this *TContext) Stop() {
	this.CloseDone()
	this.WG.Wait()
}
