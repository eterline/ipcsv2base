// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.

package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// InitLogger – create singletone style structure logger
func InitLogger(level string, json bool) {
	logger := NewLogger(level, json)
	slog.SetDefault(logger)
}

// InitLogger – create new structure logger
func NewLogger(level string, json bool) *slog.Logger {
	opt := &slog.HandlerOptions{
		Level: selectLogLevel(level),
	}

	var h slog.Handler
	if json {
		h = slog.NewJSONHandler(os.Stdout, opt)
	} else {
		h = slog.NewTextHandler(os.Stdout, opt)
	}

	return slog.New(h)
}

// Uses only for context
type loggerContextKey int

const (
	loggerCtxKey loggerContextKey = iota
)

// WrapLoggerToContext – wrap logger to parent context
func WrapLoggerToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

// LoggerFromContext – unwrap logger from context
func LoggerFromContext(ctx context.Context) (logger *slog.Logger, ok bool) {
	l, ok := ctx.Value(loggerCtxKey).(*slog.Logger)
	if ok {
		return l, true
	}
	return nil, false
}

// MustLoggerFromContext – unwrap logger from context
func MustLoggerFromContext(ctx context.Context) (logger *slog.Logger) {
	l, ok := ctx.Value(loggerCtxKey).(*slog.Logger)
	if ok {
		return l
	}
	panic("logger doesn't wrapped in this context.Context")
}

func selectLogLevel(level string) slog.Level {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
