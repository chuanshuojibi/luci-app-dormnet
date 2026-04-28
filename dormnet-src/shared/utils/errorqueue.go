package utils

import (
	"sync"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
)

type ErrQueue interface {
	Go(func() errx.Exception)
	Defer(func())
	Wait() errx.Exception
	DoDefer()
}

type errQueue struct {
	lock   *sync.Mutex
	tasks  []func() errx.Exception
	defers []func()
}

func NewErrQueue() ErrQueue {
	return &errQueue{
		lock:   &sync.Mutex{},
		tasks:  make([]func() errx.Exception, 0),
		defers: make([]func(), 0),
	}
}

func (e *errQueue) Go(block func() errx.Exception) {
	e.lock.Lock()
	e.tasks = append(e.tasks, block)
	e.lock.Unlock()
}

func (e *errQueue) Wait() errx.Exception {
	for _, task := range e.tasks {
		if err := task(); err != nil {
			return err
		}
	}
	return nil
}

func (e *errQueue) Defer(block func()) {
	e.lock.Lock()
	e.defers = append(e.defers, block)
	e.lock.Unlock()
}

func (e *errQueue) DoDefer() {
	for _, def := range e.defers {
		def()
	}
}
