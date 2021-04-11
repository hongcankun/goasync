package goasync

import "context"

// Task ...
// TODO: @hongcankun handle error
type Task func(ctx context.Context) error

type task struct {
	*options
	t Task
}

func newTask(t Task, opts *options) *task {
	return &task{
		t:       t,
		options: opts,
	}
}

func (t *task) safeRun() {
	defer func() {
		if r := recover(); r != nil {
			t.logger.Errorf("task %q panic: %v", t.taskName, r)
		}
	}()
	// TODO: @hongcankun handle error
	t.t(t.contextProvider())
}
