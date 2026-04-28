package errx4zap

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type ErrxEncoder interface {
	zapcore.Encoder
	AddRawLine(line string)
	AddRawLinef(line string, args ...interface{})
	AddException(err errx.Exception)

	CloseAllNamespace()
	CloseOneNamespace()
	CloseNamespace(depth int)
}

type EntryEncoder func(entry *zapcore.Entry, enc RawStringEncoder, cfg zapcore.EncoderConfig)

type errxEncoder struct {
	zapcore.EncoderConfig
	buffer *buffer.Buffer

	namespaceDepth int

	entryEncoder EntryEncoder
}

var _bufPool = buffer.NewPool()

func newErrxEncoder(cfg zapcore.EncoderConfig, entryEncoder EntryEncoder) ErrxEncoder {
	encoder := &errxEncoder{
		EncoderConfig:  cfg,
		buffer:         _bufPool.Get(),
		namespaceDepth: 0,
		entryEncoder:   entryEncoder,
	}
	return encoder
}

func NewErrxConsoleEncoder(cfg zapcore.EncoderConfig) ErrxEncoder {
	return newErrxEncoder(cfg, func(entry *zapcore.Entry, enc RawStringEncoder, cfg zapcore.EncoderConfig) {
		enc.AppendTime(entry.Time)
		enc.AppendString(" ")
		cfg.EncodeLevel(entry.Level, enc)
		if entry.Caller.Defined {
			enc.AppendString(" (")
			cfg.EncodeCaller(entry.Caller, enc)
			enc.AppendString(")")
		}
		enc.AppendString(": ")
		enc.AppendString(entry.Message)
	})
}

func NewErrxSyslogEncoder(cfg zapcore.EncoderConfig) ErrxEncoder {
	return newErrxEncoder(cfg, func(entry *zapcore.Entry, enc RawStringEncoder, cfg zapcore.EncoderConfig) {
		if entry.Caller.Defined {
			enc.AppendString("(")
			cfg.EncodeCaller(entry.Caller, enc)
			enc.AppendString(") ")
		}
		enc.AppendString(entry.Message)
	})
}

func (e *errxEncoder) useRawStringEncoder(key string, block func(encoder RawStringEncoder)) {
	helper := NewArr2objEncoder(key, e)
	encoder := NewExtendedRawStringEncoder(e.EncoderConfig, helper)
	block(encoder)
}

func (e *errxEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		_ = encoder.AppendArray(marshaler)
	})
	return nil
}

func (e *errxEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		_ = encoder.AppendObject(marshaler)
	})
	return nil
}

func (e *errxEncoder) AddReflected(key string, obj interface{}) error {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		_ = encoder.AppendReflected(obj)
	})
	return nil
}

func (e *errxEncoder) AddBinary(key string, bin []byte) {
	value := base64.StdEncoding.EncodeToString(bin)
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendString(value)
	})
}

func (e *errxEncoder) AddByteString(key string, value []byte) {
	e.AddString(key, string(value))
}

func (e *errxEncoder) AddBool(key string, value bool) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendBool(value)
	})
}

func (e *errxEncoder) AddComplex128(key string, value complex128) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendComplex128(value)
	})
}

func (e *errxEncoder) AddComplex64(key string, value complex64) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendComplex64(value)
	})
}

func (e *errxEncoder) AddDuration(key string, value time.Duration) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendDuration(value)
	})
}

func (e *errxEncoder) AddFloat64(key string, value float64) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendFloat64(value)
	})
}

func (e *errxEncoder) AddFloat32(key string, value float32) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendFloat32(value)
	})
}

func (e *errxEncoder) AddInt(key string, value int) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendInt(value)
	})
}

func (e *errxEncoder) AddInt64(key string, value int64) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendInt64(value)
	})
}

func (e *errxEncoder) AddInt32(key string, value int32) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendInt32(value)
	})
}

func (e *errxEncoder) AddInt16(key string, value int16) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendInt16(value)
	})
}

func (e *errxEncoder) AddInt8(key string, value int8) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendInt8(value)
	})
}

func (e *errxEncoder) AddString(key, value string) {
	e.AddRawLinef("%s: %s", key, value)
}

func (e *errxEncoder) AddTime(key string, value time.Time) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendTime(value)
	})
}

func (e *errxEncoder) AddUint(key string, value uint) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendUint(value)
	})
}

func (e *errxEncoder) AddUint64(key string, value uint64) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendUint64(value)
	})
}

func (e *errxEncoder) AddUint32(key string, value uint32) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendUint32(value)
	})
}

func (e *errxEncoder) AddUint16(key string, value uint16) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendUint16(value)
	})
}

func (e *errxEncoder) AddUint8(key string, value uint8) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendUint8(value)
	})
}

func (e *errxEncoder) AddUintptr(key string, value uintptr) {
	e.useRawStringEncoder(key, func(encoder RawStringEncoder) {
		encoder.AppendUintptr(value)
	})
}

func (e *errxEncoder) AddRawLine(line string) {
	e.addNamespacePrefix()
	_, _ = e.buffer.WriteString(line)
	_ = e.buffer.WriteByte('\n')
}

func (e *errxEncoder) AddRawLinef(line string, args ...interface{}) {
	e.AddRawLine(fmt.Sprintf(line, args...))
}

func (e *errxEncoder) AddException(err errx.Exception) {
	e.OpenNamespace(fmt.Sprintf("exception: %s", err.Error()))
	e.addExceptionWithoutHeader(err)
	e.CloseOneNamespace()
}

func (e *errxEncoder) addCause(cause errx.Exception) {
	e.OpenNamespace(fmt.Sprintf("cause by: %s", cause.Error()))
	e.addExceptionWithoutHeader(cause)
	e.CloseOneNamespace()
}

func (e *errxEncoder) addExceptionWithoutHeader(err errx.Exception) {
	if len(err.Fields()) > 0 {
		e.OpenNamespace("fields")
		for _, field := range err.Fields() {
			field.AddTo(e)
		}
		e.CloseOneNamespace()
	}

	e.OpenNamespace("stack")
	for _, frame := range err.StackTrace() {
		e.AddRawLinef("  %s:%d - %s", frame.File, frame.Line, frame.Function)
	}
	e.CloseOneNamespace()

	if err.Cause() != nil {
		e.CloseOneNamespace()
		e.addCause(err.Cause())
	}
}

func (e *errxEncoder) addNamespacePrefix() {
	for range e.namespaceDepth {
		_, _ = e.buffer.WriteString("| ")
	}
}

func (e *errxEncoder) addArg(key string, content string) {
	e.addNamespacePrefix()
	_, _ = e.buffer.WriteString(key)
	_, _ = e.buffer.WriteString(": ")
	_, _ = e.buffer.WriteString(content)
	_ = e.buffer.WriteByte('\n')
}

func (e *errxEncoder) OpenNamespace(key string) {
	e.addNamespacePrefix()
	_, _ = e.buffer.WriteString("-- ")
	_, _ = e.buffer.WriteString(key)
	_ = e.buffer.WriteByte('\n')
	e.namespaceDepth += 1
}

func (e *errxEncoder) CloseAllNamespace() {
	e.namespaceDepth = 0
}
func (e *errxEncoder) CloseOneNamespace() {
	e.CloseNamespace(1)
}
func (e *errxEncoder) CloseNamespace(depth int) {
	e.namespaceDepth -= depth
	if e.namespaceDepth < 0 {
		e.namespaceDepth = 0
	}
}

func (e *errxEncoder) Clone() zapcore.Encoder {
	newEncoder := e.clone()
	_, _ = newEncoder.(*errxEncoder).buffer.Write(e.buffer.Bytes())
	return newEncoder
}

func (e *errxEncoder) clone() ErrxEncoder {
	return newErrxEncoder(e.EncoderConfig, e.entryEncoder)
}

func (e *errxEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	enc := e.clone().(*errxEncoder)
	entryEnc := NewRawStringEncoder(enc.EncoderConfig)

	e.entryEncoder(&entry, entryEnc, entryEnc.(*rawStringEncoder).EncoderConfig)

	enc.AddRawLine(entryEnc.(*rawStringEncoder).buffer.String())

	e.OpenNamespace("args")
	for _, field := range fields {
		switch field.Type {
		case zapcore.ErrorType:
			//goland:noinspection GoTypeAssertionOnErrors
			if err, ok := field.Interface.(errx.Exception); ok {
				enc.AddException(err)
			} else {
				panic("you should always pass errx.Exception type to zap.Error()")
			}
		default:
			field.AddTo(enc)
		}
	}
	e.CloseAllNamespace()
	return enc.buffer, nil
}
