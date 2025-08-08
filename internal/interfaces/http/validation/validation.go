package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
)

// MapBindJSONError converts bind/validation errors into friendly client messages.
func MapBindJSONError(err error) (code, message string) {
	if err == nil {
		return "", ""
	}
	// Empty body
	if errors.Is(err, io.EOF) {
		return "invalid_request", "request body is empty"
	}
	// Malformed JSON
	var syn *json.SyntaxError
	if errors.As(err, &syn) {
		return "invalid_request", fmt.Sprintf("malformed JSON at position %d", syn.Offset)
	}
	// Wrong type for specific field
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		if typeErr.Field != "" {
			return "invalid_request", fmt.Sprintf("invalid type for field %s", typeErr.Field)
		}
		return "invalid_request", "invalid type in JSON payload"
	}
	// Validation tag errors
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) && len(verrs) > 0 {
		v := verrs[0]
		field := v.Field()
		switch v.Tag() {
		case "required":
			return "invalid_request", fmt.Sprintf("field %s is required", field)
		case "min":
			return "invalid_request", fmt.Sprintf("field %s must be at least %s characters", field, v.Param())
		case "email":
			return "invalid_request", fmt.Sprintf("field %s must be a valid email", field)
		default:
			return "invalid_request", fmt.Sprintf("invalid value for field %s", field)
		}
	}
	return "invalid_request", "invalid JSON payload"
}

// IsBodyTooLarge detects oversized request body errors from MaxBytesReader / JSON decoder
func IsBodyTooLarge(err error) bool {
	if err == nil {
		return false
	}
	// net/http: request body too large or similar text
	if strings.Contains(strings.ToLower(err.Error()), "request body too large") {
		return true
	}
	// json: error after MaxBytesReader may return unexpected EOF or "http: request body too large"
	if strings.Contains(strings.ToLower(err.Error()), "http: request body too large") {
		return true
	}
	return false
}
