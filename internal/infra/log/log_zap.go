package log

import (
	"os"
	"strings"

	"github.com/eterline/ipcsv2base/internal/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	log *zap.Logger
}

func NewZapLoggerWithConfig(levelStr string, dev, json, color bool) (*ZapLogger, error) {
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
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		level,
	)

	logger := zap.New(core)
	return &ZapLogger{log: logger}, nil
}

func (z *ZapLogger) With(fields ...model.LogField) model.Logger {
	return &ZapLogger{
		log: z.log.With(toZapFields(fields)...),
	}
}

func (z *ZapLogger) Debug(msg string, fields ...model.LogField) {
	z.log.Debug(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Info(msg string, fields ...model.LogField) {
	z.log.Info(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields ...model.LogField) {
	z.log.Warn(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...model.LogField) {
	z.log.Error(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Fatal(msg string, fields ...model.LogField) {
	z.log.Fatal(msg, toZapFields(fields)...)
}

func toZapFields(fields []model.LogField) []zap.Field {
	zf := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zf = append(zf, zap.Any(f.Key(), f.Value()))
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
