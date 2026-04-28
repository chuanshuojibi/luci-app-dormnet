package errx4zap

import (
	"encoding/base64"
	"time"

	"go.uber.org/zap/zapcore"
)

type ErrxJsonObjectEncoder interface {
	zapcore.ObjectEncoder
	OriginValue() map[string]any
}

type errxJsonObjectEncoder struct {
	zapcore.EncoderConfig
	items map[string]any
}

func NewErrxJsonObjectEncoder(config zapcore.EncoderConfig) ErrxJsonObjectEncoder {
	return &errxJsonObjectEncoder{
		EncoderConfig: config,
	}
}

func (e *errxJsonObjectEncoder) OriginValue() map[string]any {
	return e.items
}

func (e *errxJsonObjectEncoder) addItem(key string, value any) {
	e.items[key] = value
}

func (e *errxJsonObjectEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	encoder := NewErrxJsonArrayEncoder(e.EncoderConfig)
	_ = marshaler.MarshalLogArray(encoder)
	e.addItem(key, encoder.OriginValue())
	return nil
}

func (e *errxJsonObjectEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	encoder := NewErrxJsonObjectEncoder(e.EncoderConfig)
	_ = marshaler.MarshalLogObject(encoder)
	e.addItem(key, encoder.OriginValue())
	return nil
}

func (e *errxJsonObjectEncoder) AddBinary(key string, value []byte) {
	base64Value := base64.StdEncoding.EncodeToString(value)
	e.addItem(key, base64Value)
}

func (e *errxJsonObjectEncoder) AddByteString(key string, value []byte) {
	e.AddString(key, string(value))
}

func (e *errxJsonObjectEncoder) AddBool(key string, value bool) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddComplex128(key string, value complex128) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddComplex64(key string, value complex64) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddDuration(key string, value time.Duration) {
	encoder := NewErrxJsonArrayEncoder(e.EncoderConfig)
	e.EncodeDuration(value, encoder)
	if len(encoder.OriginValue()) >= 1 {
		e.addItem(key, encoder.OriginValue()[0])
	}
}

func (e *errxJsonObjectEncoder) AddFloat64(key string, value float64) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddFloat32(key string, value float32) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddInt(key string, value int) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddInt64(key string, value int64) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddInt32(key string, value int32) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddInt16(key string, value int16) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddInt8(key string, value int8) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddString(key, value string) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddTime(key string, value time.Time) {
	encoder := NewErrxJsonArrayEncoder(e.EncoderConfig)
	e.EncodeTime(value, encoder)
	if len(encoder.OriginValue()) >= 1 {
		e.addItem(key, encoder.OriginValue()[0])
	}
}

func (e *errxJsonObjectEncoder) AddUint(key string, value uint) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddUint64(key string, value uint64) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddUint32(key string, value uint32) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddUint16(key string, value uint16) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddUint8(key string, value uint8) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddUintptr(key string, value uintptr) {
	e.addItem(key, value)
}

func (e *errxJsonObjectEncoder) AddReflected(key string, value interface{}) error {
	e.addItem(key, value)
	return nil
}

func (e *errxJsonObjectEncoder) OpenNamespace(key string) {
	//return UnsupportedOperationException()
}
