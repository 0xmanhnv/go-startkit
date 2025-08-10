package validation

import (
	"encoding/json"
	"errors"
	"io"
	"testing"
)

func TestMapBindJSONError_EmptyBody(t *testing.T) {
	code, msg := MapBindJSONError(io.EOF)
	if code != "invalid_request" || msg == "" {
		t.Fatalf("unexpected: %s %s", code, msg)
	}
}

func TestMapBindJSONError_Syntax(t *testing.T) {
	code, msg := MapBindJSONError(&json.SyntaxError{Offset: 5})
	if code != "invalid_request" || msg == "" {
		t.Fatalf("unexpected: %s %s", code, msg)
	}
}

func TestMapBindJSONError_Type(t *testing.T) {
	code, msg := MapBindJSONError(&json.UnmarshalTypeError{Field: "age"})
	if code != "invalid_request" || msg == "" {
		t.Fatalf("unexpected: %s %s", code, msg)
	}
}

func TestIsBodyTooLarge(t *testing.T) {
	if !IsBodyTooLarge(errors.New("http: request body too large")) {
		t.Fatalf("expected true")
	}
}
