package validation

// MsgKey is a typed key for looking up message templates.
type MsgKey string

const (
	KeyFieldRequired        MsgKey = "field_required"
	KeyMinLen               MsgKey = "min_len"
	KeyEmail                MsgKey = "email"
	KeyStrongPassword       MsgKey = "strong_password"
	KeyMalformedJSONAt      MsgKey = "malformed_json_at"
	KeyInvalidTypeForField  MsgKey = "invalid_type_for_field"
	KeyInvalidValueForField MsgKey = "invalid_value_for_field"
	KeyEmptyBody            MsgKey = "empty_body"
)
