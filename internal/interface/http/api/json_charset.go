// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package api

import (
	"strings"
	"sync/atomic"
)

// CharsetStyle – string-based enum defining supported character encodings.
type CharsetStyle string

const (
	// CharsetUTF8 – UTF-8 encoding (default).
	CharsetUTF8 CharsetStyle = "utf-8"
	// CharsetUTF16 – UTF-16 encoding.
	CharsetUTF16 CharsetStyle = "utf-16"
	// CharsetUTF16LE – UTF-16 Little Endian.
	CharsetUTF16LE CharsetStyle = "utf-16le"
	// CharsetUTF16BE – UTF-16 Big Endian.
	CharsetUTF16BE CharsetStyle = "utf-16be"
	// CharsetISO88591 – ISO-8859-1 (Latin-1).
	CharsetISO88591 CharsetStyle = "iso-8859-1"
	// CharsetWindows1251 – Windows-1251 Cyrillic.
	CharsetWindows1251 CharsetStyle = "windows-1251"
	// CharsetWindows1252 – Windows-1252 Western Europe.
	CharsetWindows1252 CharsetStyle = "windows-1252"
	// CharsetGB18030 – GB18030 Simplified Chinese.
	CharsetGB18030 CharsetStyle = "gb18030"
)

/*
charsetValue – stores global charset for HTTP responses.

	Using atomic.Pointer[string] ensures safe concurrent access and avoids interface{} boxing.
*/
var charsetValue atomic.Pointer[string]

// init – package initialization, loads default charset.
func init() {
	LoadDefault()
}

/*
SetCharset – sets a new global charset and returns the previous value.

	Parameters:
		new – CharsetStyle to set globally.
	Returns:
		old – previous CharsetStyle value.
*/
func SetCharset(new CharsetStyle) (old CharsetStyle) {
	newStr := string(new)
	oldPtr := charsetValue.Swap(&newStr)
	if oldPtr == nil {
		return CharsetUTF8
	}
	return CharsetStyle(*oldPtr)
}

// CurrentCharset – returns the current global charset.
func CurrentCharset() CharsetStyle {
	ptr := charsetValue.Load()
	if ptr == nil {
		return CharsetUTF8
	}
	return CharsetStyle(*ptr)
}

// LoadDefault – sets the default global charset (UTF-8).
func LoadDefault() {
	defaultStr := string(CharsetUTF8)
	charsetValue.Store(&defaultStr)
}

/*
stringContentType – returns full Content-Type string with current charset.

	If t is empty, defaults to "application/json".
	Optimized to avoid allocations via strings.Builder.
*/
func stringContentType(t string) string {
	if t == "" {
		t = "application/json"
	}

	var b strings.Builder

	b.Grow(len(t) + len("; charset=") + len(CurrentCharset()))
	b.WriteString(t)
	b.WriteString("; charset=")
	b.WriteString(string(CurrentCharset()))
	return b.String()
}
