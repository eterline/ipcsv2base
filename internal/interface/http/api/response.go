// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseWrapQuery interface {
	SetCode(code int) ResponseWrapQuery
	SetMessage(msg string) ResponseWrapQuery

	WrapData(data any) ResponseWrapQuery
	AddError(err ...error) ResponseWrapQuery
	AddStringError(err ...string) ResponseWrapQuery

	Write(w http.ResponseWriter) error
	Raw() *ResponseHttpWrapper
}

/*
ResponseHttpWrapper – unified HTTP JSON response wrapper.

Holds status code, message, array of errors, and optional payload of type T.
Must only be used on transport level and never referenced by domain logic.
*/
type ResponseHttpWrapper struct {
	Code    int      `json:"code"`              // HTTP status or internal code
	Message string   `json:"message,omitempty"` // Optional descriptive message
	Errors  []string `json:"errors,omitempty"`  // Array of error messages
	Data    any      `json:"data,omitempty"`    // Optional payload of type T
}

// initErrs – ensures Errors slice is initialized with at least startLen capacity.
func (r *ResponseHttpWrapper) initErrs(startLen int) {
	if r.Errors == nil {
		r.Errors = make([]string, 0, startLen)
	}
}

// SetCode – sets the HTTP status code and returns the wrapper for chaining.
func (r *ResponseHttpWrapper) SetCode(code int) ResponseWrapQuery {
	r.Code = code
	return r
}

// SetMessage – sets the message and returns the wrapper for chaining.
func (r *ResponseHttpWrapper) SetMessage(msg string) ResponseWrapQuery {
	r.Message = msg
	return r
}

// WrapData – sets the payload and returns the wrapper for chaining.
func (r *ResponseHttpWrapper) WrapData(data any) ResponseWrapQuery {
	r.Data = &data
	return r
}

// AddError – adds one or more error values to the Errors slice.
func (r *ResponseHttpWrapper) AddError(err ...error) ResponseWrapQuery {
	r.initErrs(len(err))
	for _, e := range err {
		r.Errors = append(r.Errors, e.Error())
	}
	return r
}

// AddStringError – adds one or more string errors to the Errors slice.
func (r *ResponseHttpWrapper) AddStringError(err ...string) ResponseWrapQuery {
	r.initErrs(len(err))
	r.Errors = append(r.Errors, err...)
	return r
}

// AddStringError – adds one or more string errors to the Errors slice.
func (r *ResponseHttpWrapper) Raw() *ResponseHttpWrapper {
	return r
}

// ================= Builders ==================

// NewResponse – creates a new empty response wrapper.
func NewResponse() *ResponseHttpWrapper {
	return &ResponseHttpWrapper{}
}

// OkResponse – creates a 200 response with a message and no payload.
func OkResponse(message string) ResponseWrapQuery {
	return NewResponse().
		SetCode(http.StatusOK).
		SetMessage(message)
}

// OkDataResponse – creates a 200 response with payload.
func OkDataResponse[T any](data T) ResponseWrapQuery {
	return NewResponse().
		SetCode(http.StatusOK).
		WrapData(data)
}

// OkMsgResponse – creates a 200 response with both message and payload.
func OkMsgResponse[T any](message string, data T) ResponseWrapQuery {
	return NewResponse().
		SetCode(http.StatusOK).
		SetMessage(message).
		WrapData(data)
}

// ErrorSimpleResponse – creates an error response with code and string errors.
func InternalErrorResponse() ResponseWrapQuery {
	return NewResponse().
		SetCode(http.StatusInternalServerError).
		SetMessage("internal server error")
}

// ErrorSimpleResponse – creates an error response with code and string errors.
func ErrorSimpleResponse[T any](code int, errsDetails ...string) ResponseWrapQuery {
	return NewResponse().
		SetCode(code).
		SetMessage("error").
		AddStringError(errsDetails...)
}

// ErrorDetailedResponse – creates an error response with code, message, and Go errors.
func ErrorDetailedResponse[T any](code int, message string, errs ...error) ResponseWrapQuery {
	return NewResponse().
		SetCode(code).
		SetMessage(message).
		AddError(errs...)
}

// NewCustomResponse – creates a fully configurable response with string error.
func NewCustomResponse[T any](code int, message, errMsg string, data T) ResponseWrapQuery {
	return NewResponse().
		SetCode(code).
		SetMessage(message).
		AddStringError(errMsg).
		WrapData(data)
}

// NewCustomDetailedResponse – creates a fully configurable response with Go error.
func NewCustomDetailedResponse[T any](code int, message string, err error, data T) ResponseWrapQuery {
	return NewResponse().
		SetCode(code).
		SetMessage(message).
		AddError(err).
		WrapData(data)
}

// Write – writes the response as JSON into http.ResponseWriter.
// Handles status code, headers, JSON encoding and fallback on encoding error.
func (r *ResponseHttpWrapper) Write(w http.ResponseWriter) error {

	w.Header().Set("Content-Type", stringContentType(""))

	if r.Code == 0 {
		r.Code = http.StatusOK
	}

	w.WriteHeader(r.Code)

	err := json.NewEncoder(w).Encode(r)
	if err != nil {
		http.Error(
			w,
			`{"code":500,"message":"JSON encode error","errors":["failed to encode response"]`,
			http.StatusInternalServerError,
		)
	}

	if err != nil {
		return fmt.Errorf("JSON response encode error: %w", err)
	}
	return nil
}
