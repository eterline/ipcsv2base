package log

import (
	"io"
	"os"
	"strings"

	"github.com/eterline/ipcsv2base/internal/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type fieldNode struct {
	field *model.LogField
	next  *fieldNode
}

type ZapLogger struct {
	log    *zap.Logger
	fields *fieldNode
}

func NewZapLoggerWithConfigStdout(levelStr string, dev, json, color bool) (*ZapLogger, error) {
	return NewZapLoggerWithConfig(zapcore.Lock(os.Stdout), levelStr, dev, json, color)
}

func NewZapLoggerWithConfig(logWr io.Writer, levelStr string, dev, json, color bool) (*ZapLogger, error) {
	level := selectZapLevel(levelStr)

	var encoderCfg zapcore.EncoderConfig
	if dev {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

	if color {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	var encoder zapcore.Encoder
	if json {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(logWr),
		level,
	)

	logger := zap.New(core)
	return &ZapLogger{log: logger}, nil
}

func (z *ZapLogger) With(fields ...model.LogField) model.Logger {
	var head *fieldNode
	for i := len(fields) - 1; i >= 0; i-- {
		head = &fieldNode{
			field: &fields[i],
			next:  head,
		}
	}

	return &ZapLogger{
		log:    z.log,
		fields: combineFieldLists(head, z.fields),
	}
}

func (z *ZapLogger) Debug(msg string, fields ...model.LogField) {
	z.log.Debug(msg, z.buildZapFields(fields)...)
}

func (z *ZapLogger) Info(msg string, fields ...model.LogField) {
	z.log.Info(msg, z.buildZapFields(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields ...model.LogField) {
	z.log.Warn(msg, z.buildZapFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...model.LogField) {
	z.log.Error(msg, z.buildZapFields(fields)...)
}

func (z *ZapLogger) Fatal(msg string, fields ...model.LogField) {
	z.log.Fatal(msg, z.buildZapFields(fields)...)
}

func (z *ZapLogger) buildZapFields(extra []model.LogField) []zap.Field {
	count := 0
	for n := z.fields; n != nil; n = n.next {
		count++
	}
	count += len(extra)

	zf := make([]zap.Field, 0, count)

	for n := z.fields; n != nil; n = n.next {
		switch v := n.field.Value().(type) {
		case string:
			zf = append(zf, zap.String(n.field.Key(), v))
		case int:
			zf = append(zf, zap.Int(n.field.Key(), v))
		case int64:
			zf = append(zf, zap.Int64(n.field.Key(), v))
		case bool:
			zf = append(zf, zap.Bool(n.field.Key(), v))
		default:
			zf = append(zf, zap.Any(n.field.Key(), v))
		}
	}

	for i := range extra {
		switch v := extra[i].Value().(type) {
		case string:
			zf = append(zf, zap.String(extra[i].Key(), v))
		case int:
			zf = append(zf, zap.Int(extra[i].Key(), v))
		case int64:
			zf = append(zf, zap.Int64(extra[i].Key(), v))
		case bool:
			zf = append(zf, zap.Bool(extra[i].Key(), v))
		default:
			zf = append(zf, zap.Any(extra[i].Key(), v))
		}
	}

	return zf
}

func selectZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func combineFieldLists(newHead, oldHead *fieldNode) *fieldNode {
	if newHead == nil {
		return oldHead
	}
	tail := newHead
	for tail.next != nil {
		tail = tail.next
	}
	tail.next = oldHead
	return newHead
}
