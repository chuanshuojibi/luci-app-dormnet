package errx4zap

import (
	"time"

	"go.uber.org/zap/zapcore"
)

type Arr2objEncoder interface {
	zapcore.ArrayEncoder
}
type arr2objEncoder struct {
	key    string
	parent zapcore.ObjectEncoder
}

func NewArr2objEncoder(key string, parent zapcore.ObjectEncoder) Arr2objEncoder {
	encoder := &arr2objEncoder{
		key:    key,
		parent: parent,
	}
	return encoder
}

func (e *arr2objEncoder) AppendBool(value bool) {
	e.parent.AddBool(e.key, value)
}

func (e *arr2objEncoder) AppendByteString(value []byte) {
	e.parent.AddByteString(e.key, value)
}

func (e *arr2objEncoder) AppendComplex128(value complex128) {
	e.parent.AddComplex128(e.key, value)
}

func (e *arr2objEncoder) AppendComplex64(value complex64) {
	e.parent.AddComplex64(e.key, value)
}

func (e *arr2objEncoder) AppendFloat64(value float64) {
	e.parent.AddFloat64(e.key, value)
}

func (e *arr2objEncoder) AppendFloat32(value float32) {
	e.parent.AddFloat32(e.key, value)
}

func (e *arr2objEncoder) AppendInt(value int) {
	e.parent.AddInt(e.key, value)
}

func (e *arr2objEncoder) AppendInt64(value int64) {
	e.parent.AddInt64(e.key, value)
}

func (e *arr2objEncoder) AppendInt32(value int32) {
	e.parent.AddInt32(e.key, value)
}

func (e *arr2objEncoder) AppendInt16(value int16) {
	e.parent.AddInt16(e.key, value)
}

func (e *arr2objEncoder) AppendInt8(value int8) {
	e.parent.AddInt8(e.key, value)
}

func (e *arr2objEncoder) AppendString(value string) {
	e.parent.AddString(e.key, value)
}

func (e *arr2objEncoder) AppendUint(value uint) {
	e.parent.AddUint(e.key, value)
}

func (e *arr2objEncoder) AppendUint64(value uint64) {
	e.parent.AddUint64(e.key, value)
}

func (e *arr2objEncoder) AppendUint32(value uint32) {
	e.parent.AddUint32(e.key, value)
}

func (e *arr2objEncoder) AppendUint16(value uint16) {
	e.parent.AddUint16(e.key, value)
}

func (e *arr2objEncoder) AppendUint8(value uint8) {
	e.parent.AddUint8(e.key, value)
}

func (e *arr2objEncoder) AppendUintptr(value uintptr) {
	e.parent.AddUintptr(e.key, value)
}

func (e *arr2objEncoder) AppendDuration(duration time.Duration) {
	e.parent.AddDuration(e.key, duration)
}

func (e *arr2objEncoder) AppendTime(time time.Time) {
	e.parent.AddTime(e.key, time)
}

func (e *arr2objEncoder) AppendArray(marshaler zapcore.ArrayMarshaler) error {
	_ = e.parent.AddArray(e.key, marshaler)
	return nil
}

func (e *arr2objEncoder) AppendObject(marshaler zapcore.ObjectMarshaler) error {
	_ = e.parent.AddObject(e.key, marshaler)
	return nil
}

func (e *arr2objEncoder) AppendReflected(value interface{}) error {
	_ = e.parent.AddReflected(e.key, value)
	return nil
}
