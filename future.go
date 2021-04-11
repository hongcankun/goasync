package goasync

// Future ...
// TODO: @hongcankun
type Future struct {
	done chan struct{}
	err  error
	// cancel
	// listener
}

func newFuture() *Future {
	return &Future{
		done: make(chan struct{}),
		err:  nil,
	}
}

func (f *Future) Done() <-chan struct{} {
	return f.done
}

func (f *Future) Err() error {
	panic("implement me")
}
