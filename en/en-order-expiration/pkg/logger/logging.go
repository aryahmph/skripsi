package logger

import (
	"context"
	"en-order-expiration/pkg/util"
	"fmt"

	"github.com/sirupsen/logrus"
)

// Field log object
type Field struct {
	Key   string
	Value interface{}
}

// FieldFunc log
type FieldFunc func(key string, value interface{}) *Field

type Fields []Field

// NewFields create instance new field
func NewFields(p ...Field) Fields {
	x := Fields{}

	for i := 0; i < len(p); i++ {
		x.Append(p[i])
	}

	return x
}

// Append new field
func (f *Fields) Append(p Field) {
	*f = append(*f, p)
}

// Any log
func Any(k string, v interface{}) Field {
	return Field{
		Key:   k,
		Value: v,
	}
}

// EventName log
func EventName(v interface{}) Field {
	return Field{
		Key:   "name",
		Value: v,
	}
}

// EventId log
func EventId(v interface{}) Field {
	return Field{
		Key:   "id",
		Value: v,
	}
}

func EventOutputBackground(vStatusCode interface{}, vObject interface{}, vBodyRaw interface{}) Field {
	output := map[string]interface{}{
		"background": map[string]interface{}{
			"response": map[string]interface{}{
				"background_status": vStatusCode,
				"body": map[string]interface{}{
					"object": vObject,
					"raw":    vBodyRaw,
				},
			},
		},
	}
	return Field{
		Key:   "output",
		Value: output,
	}
}

func EventInputKafka(vObject interface{}, vRaw interface{}) Field {
	input := map[string]interface{}{
		"kafka": map[string]interface{}{
			"payload": map[string]interface{}{
				"object": vObject,
				"raw":    vRaw,
			},
		},
	}
	return Field{
		Key:   "input",
		Value: input,
	}
}

func EventOutputKafka(vBody interface{}, vPartition interface{}, vOffset interface{}) Field {
	output := map[string]interface{}{
		"kafka": map[string]interface{}{
			"message_body":      vBody,
			"message_partition": vPartition,
			"message_offset":    vOffset,
		},
	}
	return Field{
		Key:   "output",
		Value: output,
	}
}

func EventOutputSqs(vBody interface{}) Field {
	output := map[string]interface{}{
		"sqs": map[string]interface{}{
			"message_body": vBody,
		},
	}
	return Field{
		Key:   "output",
		Value: output,
	}
}

func EventInputHttp(v interface{}) Field {
	input := map[string]interface{}{
		"http": map[string]interface{}{
			"request": map[string]interface{}{
				"body": map[string]interface{}{
					"object": v,
				},
			},
		},
	}
	return Field{
		Key:   "input",
		Value: input,
	}
}

func EventOutputNGate(v interface{}) Field {
	return Field{
		Key:   "ngate",
		Value: v,
	}
}

func EventOutputAuthenticator(v interface{}) Field {
	return Field{
		Key:   "authenticator",
		Value: v,
	}
}

func EventInputCache(v interface{}) Field {
	return Field{
		Key:   "cache",
		Value: v,
	}
}

func EventOutputOTPConfig(v interface{}) Field {
	return Field{
		Key:   "otp_config",
		Value: v,
	}
}

func EventOutputHttp(vStatusCode interface{}, vObject interface{}, vBodyRaw interface{}) Field {
	output := map[string]interface{}{
		"http": map[string]interface{}{
			"response": map[string]interface{}{
				"http_status": vStatusCode,
				"body": map[string]interface{}{
					"object": vObject,
					"raw":    vBodyRaw,
				},
			},
		},
	}
	return Field{
		Key:   "output",
		Value: output,
	}
}

// SetMessageFormat message with custom argument
func SetMessageFormat(format string, args ...interface{}) interface{} {
	return fmt.Sprintf(format, args...)
}

func extract(args ...Field) map[string]interface{} {
	if len(args) == 0 {
		return nil
	}

	data := map[string]interface{}{}
	for _, fl := range args {
		data[fl.Key] = fl.Value
	}
	return data
}

// Error log
func Error(arg interface{}, fl ...Field) {
	logrus.WithFields(map[string]interface{}{
		"event": extract(fl...),
	}).Error(arg)
}

func Info(arg interface{}, fl ...Field) {
	logrus.WithFields(map[string]interface{}{
		"event": extract(fl...),
	}).Info(arg)
}

func Debug(arg interface{}, fl ...Field) {
	logrus.WithFields(map[string]interface{}{
		"event": extract(fl...),
	}).Debug(arg)
}

// Fatal log
func Fatal(arg interface{}, fl ...Field) {
	logrus.WithFields(map[string]interface{}{
		"event": extract(fl...),
	}).Fatal(arg)
}

// Warn log
func Warn(arg interface{}, fl ...Field) {
	logrus.WithFields(map[string]interface{}{
		"event": extract(fl...),
	}).Warn(arg)
}

// Trace log
func Trace(arg interface{}, fl ...Field) {
	logrus.WithFields(map[string]interface{}{
		"event": extract(fl...),
	}).Trace(arg)
}

// AccessLog http accessing log
func AccessLog(arg interface{}, fl ...Field) {
	logrus.WithFields(extract(fl...)).Info(arg)
}

// InfoWithContext log info with context
func InfoWithContext(ctx context.Context, arg interface{}, fl ...Field) {
	logrus.WithFields(extractContext(ctx.Value("access"), map[string]interface{}{
		"event": extract(fl...),
	})).WithContext(ctx).Info(arg)
}

// WarnWithContext log warn with context
func WarnWithContext(ctx context.Context, arg interface{}, fl ...Field) {
	logrus.WithFields(extractContext(ctx.Value("access"), map[string]interface{}{
		"event": extract(fl...),
	})).WithContext(ctx).Warn(arg)
}

// ErrorWithContext log error with context
func ErrorWithContext(ctx context.Context, arg interface{}, fl ...Field) {
	logrus.WithFields(extractContext(ctx.Value("access"), map[string]interface{}{
		"event": extract(fl...),
	})).WithContext(ctx).Error(arg)
}

// DebugWithContext log debug with context
func DebugWithContext(ctx context.Context, arg interface{}, fl ...Field) {
	logrus.WithFields(extractContext(ctx.Value("access"), map[string]interface{}{
		"event": extract(fl...),
	})).WithContext(ctx).Debug(arg)
}

// TraceWithContext log trace with context
func TraceWithContext(ctx context.Context, arg interface{}, fl ...Field) {
	logrus.WithFields(extractContext(ctx.Value("access"), map[string]interface{}{
		"event": extract(fl...),
	})).WithContext(ctx).Trace(arg)
}

func extractContext(i interface{}, logField map[string]interface{}) map[string]interface{} {
	if util.IsSameType(i, logField) {
		x := i.(map[string]interface{})
		for k, v := range x {
			logField[k] = v
		}
	}
	return logField
}
