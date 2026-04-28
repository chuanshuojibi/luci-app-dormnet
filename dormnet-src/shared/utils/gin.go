package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/log"
	"go.uber.org/zap"
)

func RunWithGin(addr string, g ErrGroup) *gin.Engine {
	r := gin.Default()
	g.Go(func() errx.Exception {
		log.Info("gin daemon started", zap.String("listen_addr", addr))
		err := r.Run(addr)
		if err != nil {
			return errx.NewExceptionWithError(err, "server stopped unexpectedly")
		}
		return nil
	})
	return r
}
