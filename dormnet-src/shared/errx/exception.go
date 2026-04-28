package errx

import (
	"fmt"
	"runtime"

	"go.uber.org/zap"
)

type LoggerFunc func(msg string, fields ...zap.Field)

type Exception interface {
	error
	Cause() Exception
	Fields() []zap.Field
	StackTrace() []runtime.Frame
}

type exception struct {
	message    string
	fields     []zap.Field
	cause      Exception
	stackTrace []runtime.Frame
}

func newException(cause Exception, msg string, args ...zap.Field) Exception {
	pc := make([]uintptr, 10)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	var stackTrace []runtime.Frame
	for {
		frame, more := frames.Next()
		stackTrace = append(stackTrace, frame)
		if !more {
			break
		}
	}
	return &exception{
		message:    msg,
		cause:      cause,
		fields:     args,
		stackTrace: stackTrace,
	}
}

func NewExceptionWithCause(cause Exception, msg string, args ...zap.Field) Exception {
	return newException(cause, msg, args...)
}

func NewExceptionOfError(err error, args ...zap.Field) Exception {
	//goland:noinspection GoTypeAssertionOnErrors
	if _, ok := err.(Exception); ok {
		panic("you should never call NewExceptionOfError on Exception type")
	}
	return newException(nil, err.Error(), args...)
}

func NewExceptionWithError(err error, msg string, args ...zap.Field) Exception {
	//goland:noinspection GoTypeAssertionOnErrors
	if _, ok := err.(Exception); ok {
		panic("you should never call NewExceptionWithError on Exception type")
	}
	return newException(nil, fmt.Sprintf("%s: %s", msg, err.Error()), args...)
}

func NewException(msg string, args ...zap.Field) Exception {
	return newException(nil, msg, args...)
}

func NewExceptionOfErrorWithCause(cause Exception, err error, args ...zap.Field) Exception {
	return newException(cause, err.Error(), args...)
}

func (e *exception) Error() string {
	return e.message
}

func (e *exception) Fields() []zap.Field {
	return e.fields
}

func (e *exception) Cause() Exception {
	return e.cause
}

func (e *exception) StackTrace() []runtime.Frame {
	return e.stackTrace
}
