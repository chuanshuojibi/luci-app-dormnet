package utils

import (
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"golang.org/x/sync/errgroup"
)

type ErrGroup interface {
	Go(func() errx.Exception)
	Wait() errx.Exception
}

type errGroup struct {
	g *errgroup.Group
}

func NewErrGroup() ErrGroup {
	var g errgroup.Group

	errg := &errGroup{
		g: &g,
	}

	return errg
}

func (e *errGroup) Go(block func() errx.Exception) {
	e.g.Go(func() error {
		return block()
	})
}

func (e *errGroup) Wait() errx.Exception {
	if err := e.g.Wait(); err != nil {
		return err.(errx.Exception)
	}
	return nil
}
