package errx4zap

import (
	"reflect"
	"time"

	"go.uber.org/zap/zapcore"
)

type ErrxJsonArrayEncoder interface {
	zapcore.ArrayEncoder
	OriginValue() []any
}

type errxJsonArrayEncoder struct {
	zapcore.EncoderConfig
	items []any
}

func NewErrxJsonArrayEncoder(config zapcore.EncoderConfig) ErrxJsonArrayEncoder {
	return &errxJsonArrayEncoder{
		EncoderConfig: config,
	}
}

func (e *errxJsonArrayEncoder) OriginValue() []any {
	return e.items
}

func (e *errxJsonArrayEncoder) AppendArray(marshaler zapcore.ArrayMarshaler) error {
	encoder := NewErrxJsonArrayEncoder(e.EncoderConfig)
	_ = marshaler.MarshalLogArray(encoder)
	e.items = append(e.items, encoder.OriginValue())
	return nil
}

func (e *errxJsonArrayEncoder) AppendObject(marshaler zapcore.ObjectMarshaler) error {
	encoder := NewErrxJsonObjectEncoder(e.EncoderConfig)
	_ = marshaler.MarshalLogObject(encoder)
	e.items = append(e.items, encoder.OriginValue())
	return nil
}

func (e *errxJsonArrayEncoder) AppendReflected(value interface{}) error {
	if reflect.ValueOf(value).Kind() == reflect.Slice {
		e.items = append(e.items, value)
	}
	return nil
}

func (e *errxJsonArrayEncoder) addItem(item any) {
	e.items = append(e.items, item)
}

func (e *errxJsonArrayEncoder) AppendBool(value bool) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendByteString(value []byte) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendComplex128(value complex128) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendComplex64(value complex64) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendFloat64(value float64) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendFloat32(value float32) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendInt(value int) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendInt64(value int64) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendInt32(value int32) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendInt16(value int16) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendInt8(value int8) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendString(value string) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendUint(value uint) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendUint64(value uint64) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendUint32(value uint32) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendUint16(value uint16) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendUint8(value uint8) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendUintptr(value uintptr) {
	e.addItem(value)
}
func (e *errxJsonArrayEncoder) AppendDuration(value time.Duration) {
	e.EncodeDuration(value, e)
}
func (e *errxJsonArrayEncoder) AppendTime(value time.Time) {
	e.EncodeTime(value, e)
}
