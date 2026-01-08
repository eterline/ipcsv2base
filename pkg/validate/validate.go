package validate

import (
	"fmt"
	"strings"
)

type ValidationErrorWrapper struct {
	Errors map[string]string
}

func NewValidationErrorWrapper() *ValidationErrorWrapper {
	return &ValidationErrorWrapper{Errors: make(map[string]string)}
}

func (v *ValidationErrorWrapper) Error() string {
	var sb strings.Builder
	for field, msg := range v.Errors {
		sb.WriteString(fmt.Sprintf("%s: %s; ", field, msg))
	}
	return strings.TrimSpace(sb.String())
}
