package goasync

import (
	"sync/atomic"
	"unsafe"
)

type worker struct {
	name     string
	executor *Executor
}

func newWorker(name string, e *Executor) *worker {
	return &worker{
		name:     name,
		executor: e,
	}
}

const (
	removeReasonDone = "done"
)

func (w *worker) run() {
	for {
		select {
		case t := <-w.executor.tasks:
			t.safeRun()
		case <-w.executor.done:
			w.executor.removeWorker(w, removeReasonDone)
			return
		}
	}
}

func (w *worker) dispose() bool {
	swapped := atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&w.executor)),
		unsafe.Pointer(w.executor), unsafe.Pointer(nil))
	if !swapped {
		return false
	}
	return true
}
