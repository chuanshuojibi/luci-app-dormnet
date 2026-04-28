package utils

import (
	"time"

	"github.com/openwrt-dormnet/dormnet/shared/log"
	"go.uber.org/zap"
)

func DelayAndLog(duration time.Duration) {
	log.Info("sleep for a while...", zap.Duration("duration", duration))
	<-time.After(duration)
}

func DelayJoinTasks(times int, delay time.Duration, block func(index int) bool) {
	for i := 0; i < times; i++ {
		if block(i) {
			break
		}
		if i < times-1 {
			<-time.After(delay)
		}
	}
}
