package utils

import (
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"go.uber.org/zap"
)

func CheckExecutableExist(name ...string) errx.Exception {
	cli, err := NewShellCli()
	if err != nil {
		return errx.NewExceptionWithCause(err, "failed to start executable check")
	}
	defer cli.Close()
	for _, exec := range name {
		code, _, err := cli.Run("which", exec)
		if err != nil {
			return errx.NewExceptionWithCause(err, "failed to check executable", zap.String("exec", exec))
		}
		if code != 0 {
			return errx.NewExceptionWithCause(err, "executable not exist", zap.String("exec", exec))
		}
		log.Debug("executable exists", zap.String("exec", exec))
	}
	return nil
}
