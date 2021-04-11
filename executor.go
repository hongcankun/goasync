package goasync

import (
	"errors"
	"sync/atomic"
	"time"
)

var (
	// ErrSubmitTimeout ...
	ErrSubmitTimeout = errors.New("submit timeout")
)

// State ...
type State = uint64

const (
	// StateInitialized ...
	StateInitialized State = iota
	// StateExecuting ...
	StateExecuting
	// StateTerminating ...
	StateTerminating
	// StateTerminated ...
	StateTerminated
)

// Executor ...
type Executor struct {
	*options
	Name string

	state     State
	numWorker uint64
	idxWorker uint64

	tasks chan *task
	done  chan struct{}
}

// NewExecutor ...
func NewExecutor(name string, opts ...Option) (*Executor, error) {
	o := newOptions()
	err := o.apply(opts...)
	if err != nil {
		return nil, err
	}

	e := &Executor{
		Name:    name,
		options: o,
		state:   StateInitialized,
		tasks:   make(chan *task, o.backlog),
		done:    make(chan struct{}),
	}
	e.start()
	return e, nil
}

// Submit ...
func (e *Executor) Submit(t Task, opts ...Option) (Future, error) {
	o := e.options.clone()
	err := o.apply(opts...)
	if err != nil {
		return nil, err
	}

	// TODO: @hongcankun return future
	return nil, e.submit(newTask(t, o))
}

func (e *Executor) submit(t *task) error {
	// TODO: @hongcankun check state
	select {
	case e.tasks <- t:
		return nil
	case <-time.After(t.submitTimeout):
		return ErrSubmitTimeout
	}
}

// State ...
func (e *Executor) State() State {
	return atomic.LoadUint64(&e.state)
}

func (e *Executor) changeState(before, after State) bool {
	return atomic.CompareAndSwapUint64(&e.state, before, after)
}

func (e *Executor) start() {
	if !e.changeState(StateInitialized, StateExecuting) {
		return
	}

	e.logger.Infof("ready to start %d workers for executor %s", e.maxConcurrency, e.Name)
	for i := uint64(0); i < e.maxConcurrency; i++ {
		e.addWorker()
	}
	e.logger.Infof("executor %s has started", e.Name)
	return
}

// Terminate ...
func (e *Executor) Terminate(wait time.Duration) {
	if e.State() == StateTerminated || e.State() == StateTerminating {
		return
	}

	// TODO: @hongcankun terminate gracefully
	if wait <= 0 {

	}
	e.changeState(StateExecuting, StateTerminating)

}

func (e *Executor) addWorker() uint64 {
	idx := atomic.AddUint64(&e.idxWorker, 1)
	w := newWorker(e.workerNameGenerator(e, idx), e)
	go w.run()
	num := atomic.AddUint64(&e.numWorker, 1)
	e.logger.Infof("a new worker %s is added to executor %s: now=%d", w.name, e.Name, num)
	return num
}

func (e *Executor) removeWorker(w *worker, reason string) uint64 {
	ok := w.dispose()
	if !ok {
		e.logger.Errorf("failed to remove worker %s", w.name)
		return atomic.LoadUint64(&e.numWorker)
	}

	num := atomic.AddUint64(&e.numWorker, -1)
	e.logger.Infof("worker %s is removed for reason %q: now=%d", reason, num)
	return num
}

func (e *Executor) dispose() bool {
	if e.State() != StateTerminated {
		e.logger.Errorf("executor %s has not terminated, can not be disposed", e.Name)
		return false
	}

	cnt := 0
	for t := range e.tasks {
		// drain all tasks
		e.logger.Warnf("task %q is drained", t.taskName)
		cnt++
	}
	e.logger.Warnf("executor %s has %d tasks been drained", e.Name, cnt)
	e.tasks = nil
	e.done = nil

	return true
}

// TODO: @hongcankun support once & fork-join
