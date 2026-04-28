package log

import (
	"log/syslog"
	"os"
	"strconv"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"github.com/openwrt-dormnet/dormnet/shared/errx/errx4zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var rootLogger *zap.Logger
var sysLogger *syslog.Writer

var Debug func(msg string, fields ...zap.Field)
var Info func(msg string, fields ...zap.Field)
var Warn func(msg string, fields ...zap.Field)
var Error func(msg string, fields ...zap.Field)
var Fatal func(msg string, fields ...zap.Field)

func RootLogger() *zap.Logger {
	return rootLogger
}

func InitLogger() errx.Exception {
	cores := make([]zapcore.Core, 0)

	if value, exist := os.LookupEnv("DORMNET_ENABLE_CONSOLE"); exist {
		if value, err := strconv.ParseBool(value); err == nil && value {
			cfg := zap.NewDevelopmentEncoderConfig()
			cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
			cores = append(cores, zapcore.NewCore(
				errx4zap.NewErrxSyslogEncoder(cfg), // TODO: switch to JSONEncoder
				zapcore.AddSync(os.Stdout),
				zapcore.DebugLevel))
		}
	}

	var err error
	sysLogger, err = syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "dormnet")
	if err == nil {
		cores = append(cores, zapcore.NewCore(
			errx4zap.NewErrxSyslogEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(sysLogger),
			zapcore.InfoLevel))
	}

	rootLogger = zap.New(zapcore.NewTee(cores...), zap.AddCaller()).Named("main")

	if err != nil {
		return errx.NewExceptionWithError(err, "failed to initialize syslog")
	}

	Debug = rootLogger.Debug
	Info = rootLogger.Info
	Warn = rootLogger.Warn
	Error = rootLogger.Error
	Fatal = rootLogger.Fatal

	return nil
}

func Sync() {
	if rootLogger != nil {
		_ = rootLogger.Sync()
	}
	if sysLogger != nil {
		sysLogger.Close()
	}
}
