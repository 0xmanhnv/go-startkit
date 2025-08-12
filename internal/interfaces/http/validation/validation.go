package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	appval "gostartkit/pkg/validator"

	"github.com/gin-gonic/gin/binding"

	resp "gostartkit/internal/interfaces/http/response"

	"github.com/go-playground/validator/v10"
)

// FieldError represents a single field validation error.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// tagFormatter formats a single validator.FieldError into (code, message).
type tagFormatter func(fe validator.FieldError) (code, message string)

var (
	formatterRegistry = map[string]tagFormatter{}
)

// RegisterTagFormatter allows customizing messages per validator tag.
func RegisterTagFormatter(tag string, fn tagFormatter) { formatterRegistry[strings.ToLower(tag)] = fn }

func init() {
	// Default formatters
	RegisterTagFormatter("required", func(fe validator.FieldError) (string, string) {
		return resp.CodeInvalidRequest, MsgFieldRequired(fe.Field())
	})
	RegisterTagFormatter("min", func(fe validator.FieldError) (string, string) {
		return resp.CodeInvalidRequest, MsgMinLen(fe.Field(), fe.Param())
	})
	RegisterTagFormatter("email", func(fe validator.FieldError) (string, string) {
		return resp.CodeInvalidRequest, MsgEmail(fe.Field())
	})
	RegisterTagFormatter("strict_email", func(fe validator.FieldError) (string, string) {
		return resp.CodeInvalidRequest, MsgEmail(fe.Field())
	})
	RegisterTagFormatter("strong_password", func(fe validator.FieldError) (string, string) {
		return resp.CodeInvalidRequest, MsgStrongPassword(fe.Field())
	})

	// Register custom validators with Gin's validator engine so tags can be used in DTOs
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Use json tag names in error fields instead of struct field names
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			if name == "" {
				return fld.Name
			}
			return name
		})
		_ = v.RegisterValidation("strict_email", StrictEmailValidator())
		_ = v.RegisterValidation("strong_password", StrongPasswordValidator())
	}
}

// MapBindJSONError converts bind/validation errors into friendly client messages (first error only).
func MapBindJSONError(err error) (code, message string) {
	if err == nil {
		return "", ""
	}
	// Empty body
	if errors.Is(err, io.EOF) {
		return resp.CodeInvalidRequest, "request body is empty"
	}
	// Malformed JSON
	var syn *json.SyntaxError
	if errors.As(err, &syn) {
		return resp.CodeInvalidRequest, MsgMalformedJSONAt(syn.Offset)
	}
	// Wrong type for specific field
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		if typeErr.Field != "" {
			return resp.CodeInvalidRequest, MsgInvalidTypeForField(typeErr.Field)
		}
		return resp.CodeInvalidRequest, resp.MsgInvalidJSON
	}
	// Validation tag errors (first one)
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) && len(verrs) > 0 {
		v := verrs[0]
		if fn, ok := formatterRegistry[strings.ToLower(v.Tag())]; ok {
			return fn(v)
		}
		return resp.CodeInvalidRequest, MsgInvalidValueForField(v.Field())
	}
	return resp.CodeInvalidRequest, resp.MsgInvalidJSON
}

// MapBindJSONErrorWithLocale is MapBindJSONError but messages rendered per locale.
func MapBindJSONErrorWithLocale(locale string, err error) (code, message string) {
	if err == nil {
		return "", ""
	}
	if errors.Is(err, io.EOF) {
		return resp.CodeInvalidRequest, MsgEmptyBody() // could localize later if needed
	}
	var syn *json.SyntaxError
	if errors.As(err, &syn) {
		return resp.CodeInvalidRequest, renderKeyLocale(locale, KeyMalformedJSONAt, "", fmt.Sprintf("%d", syn.Offset))
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		if typeErr.Field != "" {
			return resp.CodeInvalidRequest, renderKeyLocale(locale, KeyInvalidTypeForField, typeErr.Field, "")
		}
		return resp.CodeInvalidRequest, resp.MsgInvalidJSON
	}
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) && len(verrs) > 0 {
		// defer to validation errors localization path
		fes := MapValidationErrorsWithLocale(locale, err)
		if len(fes) > 0 {
			return resp.CodeInvalidRequest, fes[0].Message
		}
	}
	return resp.CodeInvalidRequest, resp.MsgInvalidJSON
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

// StrictEmailValidator returns a go-playground/validator Func that enforces strict email via pkg/validator.
func StrictEmailValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		return appval.IsValidEmail(v)
	}
}

// StrongPasswordValidator checks baseline strong password rules via pkg/validator.
func StrongPasswordValidator() validator.Func {
	return func(fl validator.FieldLevel) bool {
		v := fl.Field().String()
		return appval.IsStrongPassword(v)
	}
}

// MapValidationErrors maps all validation errors into a list of FieldError for richer client responses.
func MapValidationErrors(err error) []FieldError {
	var out []FieldError
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) || len(verrs) == 0 {
		return out
	}
	for _, fe := range verrs {
		code, msg := func() (string, string) {
			if fn, ok := formatterRegistry[strings.ToLower(fe.Tag())]; ok {
				return fn(fe)
			}
			return resp.CodeInvalidRequest, fmt.Sprintf("invalid value for field %s", fe.Field())
		}()
		out = append(out, FieldError{Field: fe.Field(), Code: code, Message: msg})
	}
	return out
}

// MapValidationErrorsWithLocale renders validator errors using locale-aware messages.
func MapValidationErrorsWithLocale(locale string, err error) []FieldError {
	var out []FieldError
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) || len(verrs) == 0 {
		return out
	}
	for _, fe := range verrs {
		tag := strings.ToLower(fe.Tag())
		field := fe.Field()
		var msg string
		switch tag {
		case "required":
			msg = renderKeyLocale(locale, KeyFieldRequired, field, "")
		case "min":
			msg = renderKeyLocale(locale, KeyMinLen, field, fe.Param())
		case "email", "strict_email":
			msg = renderKeyLocale(locale, KeyEmail, field, "")
		case "strong_password":
			msg = renderKeyLocale(locale, KeyStrongPassword, field, "")
		default:
			msg = renderKeyLocale(locale, KeyInvalidValueForField, field, "")
		}
		out = append(out, FieldError{Field: field, Code: resp.CodeInvalidRequest, Message: msg})
	}
	return out
}
