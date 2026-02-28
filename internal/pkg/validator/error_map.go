package validator

import (
	"fmt"
	"strings"
)

type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

type MapValidationError struct {
	Errors map[string]error
}

func NewErrorMap(errs map[string]error) error {
	return &MapValidationError{Errors: errs}
}

func (e MapValidationError) Error() string {
	errorMessage := make([]string, 0)

	for key, er := range e.Errors {
		errorMessage = append(errorMessage, fmt.Sprintf("%s:%s", key, er.Error()))
	}
	return strings.Join(errorMessage, ";")
}
