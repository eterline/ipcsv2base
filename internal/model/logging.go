package model

import (
	"fmt"
	"strconv"
	"strings"
)

// LogField - represents a single structured log field as a key–value pair.
type LogField struct {
	key   string
	value any
}

func (f LogField) Key() string {
	return f.key
}

func (f LogField) Value() any {
	return f.value
}

// FieldError - creates a LogField with error.
func FieldError(err error) LogField {
	return FieldErrorKey("error", err)
}

// FieldErrorKey - creates a LogField with error.
func FieldErrorKey(key string, err error) LogField {
	return FieldString(key, err.Error())
}

// FieldString - creates a LogField with a string value.
func FieldString(key, v string) LogField {
	return LogField{key: key, value: v}
}

// FieldStringJoin - creates a LogField with a strings joined with joinSep.
func FieldStringJoin(key, joinSep string, v ...string) LogField {
	return FieldString(key, strings.Join(v, joinSep))
}

// FieldFloat - creates a LogField from a float64 value formatted
// with the given number of decimal places.
func FieldFloat(key string, v float64, prec int) LogField {
	return FieldString(key, strconv.FormatFloat(v, 'f', prec, 64))
}

// FieldStringer - creates a LogField from a value implementing fmt.Stringer.
// If the provided value is nil, the string "nil" is used as the field value.
func FieldStringer(key string, v fmt.Stringer) LogField {
	if v == nil {
		return LogField{key: key, value: "nil"}
	}
	return LogField{key: key, value: v}
}

// Field - creates a LogField with an arbitrary value.
func Field(key string, v any) LogField {
	return LogField{key: key, value: v}
}

/*
Logger - interface for a structured logger.

	Provides methods for attaching contextual fields and emitting log messages
	at different severity levels.
*/
type Logger interface {
	// With - returns a new Logger instance with additional contextual fields attached.
	With(fields ...LogField) Logger

	// Debug - logs a message at Debug level.
	Debug(msg string, fields ...LogField)

	// Info - logs a message at Info level.
	Info(msg string, fields ...LogField)

	// Warn - logs a message at Warn level.
	Warn(msg string, fields ...LogField)

	// Error - logs a message at Error level.
	Error(msg string, fields ...LogField)

	// Fatal - logs a message at Fatal level and terminates the application.
	Fatal(msg string, fields ...LogField)
}
