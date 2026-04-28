package errx4zap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type RawStringEncoder interface {
	zapcore.ArrayEncoder
}

type rawStringEncoder struct {
	zapcore.EncoderConfig
	parent zapcore.ArrayEncoder
	buffer *buffer.Buffer
}

func NewExtendedRawStringEncoder(config zapcore.EncoderConfig, parent zapcore.ArrayEncoder) RawStringEncoder {
	return &rawStringEncoder{
		EncoderConfig: config,
		parent:        parent,
	}
}

func NewRawStringEncoder(config zapcore.EncoderConfig) RawStringEncoder {
	return &rawStringEncoder{
		EncoderConfig: config,
		buffer:        _bufPool.Get(),
	}
}

func (e *rawStringEncoder) writeString(value string) {
	if e.parent != nil {
		e.parent.AppendString(value)
	} else if e.buffer != nil {
		_, _ = e.buffer.WriteString(value)
	}
}

func (e *rawStringEncoder) writeByteString(value []byte) {
	if e.parent != nil {
		e.parent.AppendByteString(value)
	} else if e.buffer != nil {
		_, _ = e.buffer.Write(value)
	}
}

func (e *rawStringEncoder) AppendBool(b bool) {
	e.writeString(strconv.FormatBool(b))
}

func (e *rawStringEncoder) AppendByteString(bytes []byte) {
	e.writeByteString(bytes)
}

func (e *rawStringEncoder) AppendComplex128(c complex128) {
	e.AppendString(strconv.FormatComplex(c, 'g', -1, 128))
}

func (e *rawStringEncoder) AppendComplex64(c complex64) {
	e.AppendComplex128(complex128(c))
}

func (e *rawStringEncoder) AppendFloat64(f float64) {
	e.writeString(strconv.FormatFloat(f, 'g', -1, 64))
}

func (e *rawStringEncoder) AppendFloat32(f float32) {
	e.AppendFloat64(float64(f))
}

func (e *rawStringEncoder) AppendInt(i int) {
	e.writeString(strconv.Itoa(i))
}

func (e *rawStringEncoder) AppendInt64(i int64) {
	e.writeString(strconv.FormatInt(i, 10))
}

func (e *rawStringEncoder) AppendInt32(i int32) {
	e.AppendInt(int(i))
}

func (e *rawStringEncoder) AppendInt16(i int16) {
	e.AppendInt(int(i))
}

func (e *rawStringEncoder) AppendInt8(i int8) {
	e.AppendInt(int(i))
}

func (e *rawStringEncoder) AppendString(s string) {
	e.writeString(s)
}

func (e *rawStringEncoder) AppendUint(u uint) {
	e.AppendUint64(uint64(u))
}

func (e *rawStringEncoder) AppendUint64(u uint64) {
	e.writeString(strconv.FormatUint(u, 10))
}

func (e *rawStringEncoder) AppendUint32(u uint32) {
	e.AppendUint64(uint64(u))
}

func (e *rawStringEncoder) AppendUint16(u uint16) {
	e.AppendUint64(uint64(u))
}

func (e *rawStringEncoder) AppendUint8(u uint8) {
	e.AppendUint64(uint64(u))
}

func (e *rawStringEncoder) AppendUintptr(u uintptr) {
	e.writeString(fmt.Sprintf("0x%x", u))
}

func (e *rawStringEncoder) AppendDuration(duration time.Duration) {
	e.EncodeDuration(duration, e)
}

func (e *rawStringEncoder) AppendTime(time time.Time) {
	e.EncodeTime(time, e)
}

func (e *rawStringEncoder) AppendArray(marshaler zapcore.ArrayMarshaler) error {
	encoder := NewErrxJsonArrayEncoder(e.EncoderConfig)
	_ = marshaler.MarshalLogArray(encoder)
	value, _ := json.Marshal(encoder.OriginValue())
	e.writeByteString(value)
	return nil
}

func (e *rawStringEncoder) AppendObject(marshaler zapcore.ObjectMarshaler) error {
	encoder := NewErrxJsonObjectEncoder(e.EncoderConfig)
	_ = marshaler.MarshalLogObject(encoder)
	value, _ := json.Marshal(encoder.OriginValue())
	e.writeByteString(value)
	return nil
}

func (e *rawStringEncoder) AppendReflected(obj interface{}) error {
	value, _ := json.Marshal(obj)
	e.writeByteString(value)
	return nil
}
