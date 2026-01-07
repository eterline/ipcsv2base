// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
)

/*
formExtractor is responsible for extracting multipart/form-data from an incoming HTTP request.

	It encapsulates:
	- request body size limiting
	- multipart parsing
	- file extraction by form key

	The extractor is intended to be used at the API / transport layer
	and assumes exclusive ownership of the request body.
*/
type formExtractor struct {
	limit int64  // Maximum allowed request body size in bytes
	key   string // Multipart form field key used to locate the file
}

/*
NewFormExtractor creates a new formExtractor instance.

	If the provided limit is less than or equal to zero,
	a default limit of 10 MB is applied.

	The key parameter specifies the multipart form field
	from which the file will be extracted.
*/
func NewFormExtractor(limit int64, key string) *formExtractor {
	max := limit
	if max <= 0 {
		max = 1 << 20 * 10
	}

	return &formExtractor{
		limit: max,
		key:   key,
	}
}

/*
wrapLimit applies an HTTP request body size limit using http.MaxBytesReader.

	This function mutates the incoming *http.Request
	and must be called before any reads from r.Body occur.
*/
func (fr *formExtractor) wrapLimit(r *http.Request) {
	if fr.limit > 0 {
		r.Body = http.MaxBytesReader(nil, r.Body, fr.limit)
	}
}

/*
FormStream extracts a file from a multipart/form-data request and returns it as an io.ReadCloser.

	Ownership of the returned stream is transferred to the caller,
	who is responsible for closing it.
*/
func (fr *formExtractor) FormStream(r *http.Request) (io.ReadCloser, error) {
	fr.wrapLimit(r)

	err := r.ParseMultipartForm(fr.limit)
	if err != nil {
		return nil, err
	}

	file, _, err := r.FormFile(fr.key)
	if err != nil {
		return nil, err
	}

	return file, nil
}

/*
FormBytes extracts a file from a multipart/form-data request and reads it entirely into memory as a byte slice.

	This method internally uses FormStream and guarantees
	that the returned stream is properly closed.
*/
func (fr *formExtractor) FormBytes(r *http.Request) ([]byte, error) {
	stream, err := fr.FormStream(r)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	return io.ReadAll(stream)
}

/*
Validatable represents a DTO that can validate its own contents.

	This interface is intended for API-layer data transfer objects
	and is typically used by extractors to enforce input validation.
*/
type Validatable interface {
	Validate() error
}

/*
ExtractJSON reads and decodes a JSON request body into a value of type T and then validates the decoded value using its Validate method.

	This function assumes ownership of the request body and always closes it.
	It must be the only component responsible for reading r.Body.

	If decoding or validation fails, a zero value of T is returned
	along with a wrapped error.
*/
func ExtractJSON[T Validatable](r *http.Request) (T, error) {
	var value T

	if err := requireJSON(r); err != nil {
		return value, fmt.Errorf("invalid request content type: %w", err)
	}

	defer r.Body.Close()

	decd := json.NewDecoder(r.Body)
	decd.DisallowUnknownFields()

	if err := decd.Decode(&value); err != nil {
		return value, fmt.Errorf("decode JSON error: %w", err)
	}

	if decd.More() {
		return value, errors.New("unexpected extra JSON data")
	}

	if err := value.Validate(); err != nil {
		return value, fmt.Errorf("extract JSON validation error: %w", err)
	}

	return value, nil
}

func requireJSON(r *http.Request) error {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return errors.New("missing Content-Type header")
	}

	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return fmt.Errorf("invalid Content-Type: %w", err)
	}

	if mediaType != "application/json" {
		return fmt.Errorf("unsupported Content-Type: %s", mediaType)
	}

	return nil
}
