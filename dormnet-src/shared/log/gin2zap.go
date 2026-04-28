package log

import (
	"strings"

	"go.uber.org/zap"
)

type SimpleLineWriter struct {
	logger *zap.Logger
}

func NewSimpleLineWriter(logger *zap.Logger) *SimpleLineWriter {
	return &SimpleLineWriter{logger: logger}
}

func (slw *SimpleLineWriter) Write(p []byte) (int, error) {
	line := strings.TrimRight(string(p), "\n")
	slw.logger.Info(line)
	return len(p), nil
}
