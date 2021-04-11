package goasync

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// ErrOptNilContextProvider ...
	ErrOptNilContextProvider = errors.New("nil context provider")
	// ErrOptInvalidConcurrency ...
	ErrOptInvalidConcurrency = errors.New("invalid concurrency")
	// ErrOptNilWorkerNameGenerator ...
	ErrOptNilWorkerNameGenerator = errors.New("nil worker name generator")
	// ErrOptNilLogger ...
	ErrOptNilLogger = errors.New("nil logger")
	// ErrOptInvalidTimeout ...
	ErrOptInvalidTimeout = errors.New("invalid timeout")
	// ErrOptNilContext ...
	ErrOptNilContext = errors.New("nil context")
)

var (
	// OptUnlimitedConcurrency ...
	OptUnlimitedConcurrency = 0
)

// WithTaskName ...
func WithTaskName(name string) Option {
	return &option{
		applyFunc: func(opts *options) {
			opts.taskName = name
		},
		validateFunc: nil,
	}
}

// WithSubmitTimeout ...
func WithSubmitTimeout(t time.Duration) Option {
	return &option{
		applyFunc: func(opts *options) {
			opts.submitTimeout = t
		},
		validateFunc: func() error {
			if t <= 0 {
				return ErrOptInvalidTimeout
			}
			return nil
		},
	}
}

// WithBacklog ...
func WithBacklog(n uint64) Option {
	return &option{
		applyFunc: func(opts *options) {
			opts.backlog = n
		},
		validateFunc: nil,
	}
}

// WithMaxConcurrency ...
func WithMaxConcurrency(c uint64) Option {
	return &option{
		applyFunc: func(opts *options) {
			opts.maxConcurrency = c
		},
		validateFunc: nil,
	}
}

// WithContextProvider ...
func WithContextProvider(p func() context.Context) Option {
	return &option{
		applyFunc: func(opts *options) {
			opts.contextProvider = p
		},
		validateFunc: func() error {
			if p == nil {
				return ErrOptNilContextProvider
			}
			return nil
		},
	}
}

// WithContext ...
func WithContext(ctx context.Context) Option {
	return &option{
		applyFunc: func(opts *options) {
			opts.contextProvider = func() context.Context {
				return ctx
			}
		},
		validateFunc: func() error {
			if ctx == nil {
				return ErrOptNilContext
			}
			return nil
		},
	}
}

// WithWorkerNameGenerator ...
func WithWorkerNameGenerator(f func(e *Executor, idx uint64) string) Option {
	return &option{
		applyFunc: func(opts *options) {
			opts.workerNameGenerator = f
		},
		validateFunc: func() error {
			if f == nil {
				return ErrOptNilWorkerNameGenerator
			}
			return nil
		},
	}
}

// WithLogger ...
func WithLogger(l Logger) Option {
	return &option{
		applyFunc: func(opts *options) {
			opts.logger = l
		},
		validateFunc: func() error {
			if l == nil {
				return ErrOptNilLogger
			}
			return nil
		},
	}
}

// Option ...
type Option interface {
	apply(opts *options)
	validate() error
}

type option struct {
	applyFunc    func(opts *options)
	validateFunc func() error
}

func (o *option) apply(opts *options) {
	o.applyFunc(opts)
}

func (o *option) validate() error {
	if o.validateFunc == nil {
		return nil
	}
	return o.validateFunc()
}

type options struct {
	maxConcurrency uint64
	backlog        uint64

	contextProvider     func() context.Context
	workerNameGenerator func(e *Executor, idx uint64) string

	logger Logger

	submitTimeout time.Duration

	// task specific options
	taskName string
}

func newOptions() *options {
	numCPU := uint64(runtime.NumCPU())
	return &options{
		maxConcurrency: numCPU,
		backlog:        numCPU * 2,
		contextProvider: func() context.Context {
			return context.Background()
		},
		workerNameGenerator: func(e *Executor, idx uint64) string {
			return fmt.Sprintf("executor-%s-worker-%d", e.Name, idx)
		},
		logger:        logrus.StandardLogger(),
		submitTimeout: time.Second * 2,
		taskName:      "<unknown>",
	}
}

func (o *options) apply(opts ...Option) error {
	for _, opt := range opts {
		err := opt.validate()
		if err != nil {
			return err
		}
		opt.apply(o)
	}
	return nil
}

func (o *options) clone() *options {
	cloned := *o
	return &cloned
}
