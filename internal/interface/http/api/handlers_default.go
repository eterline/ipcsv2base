// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package api

import (
	"fmt"
	"net/http"
)

func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	NewResponse().
		SetCode(http.StatusNotFound).
		SetMessage("handler not found").
		Write(w)
}

func HandleNotAllowedMethod(w http.ResponseWriter, r *http.Request) {
	NewResponse().
		SetCode(http.StatusMethodNotAllowed).
		SetMessage(fmt.Sprintf("%s: method not allowed", r.Method)).
		Write(w)
}
