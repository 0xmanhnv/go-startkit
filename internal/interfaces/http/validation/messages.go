package validation

import (
	"gostartkit/pkg/i18n"
	"fmt"
)

// Message helpers (single source of truth)

func MsgFieldRequired(field string) string { return renderKey(KeyFieldRequired, field, "") }

func MsgMinLen(field string, min string) string { return renderKey(KeyMinLen, field, min) }

func MsgEmail(field string) string { return renderKey(KeyEmail, field, "") }

func MsgStrongPassword(field string) string { return renderKey(KeyStrongPassword, field, "") }

func MsgMalformedJSONAt(offset int64) string {
	return renderKey(KeyMalformedJSONAt, "", fmt.Sprintf("%d", offset))
}

func MsgInvalidTypeForField(field string) string { return renderKey(KeyInvalidTypeForField, field, "") }

func MsgInvalidValueForField(field string) string {
	return renderKey(KeyInvalidValueForField, field, "")
}

func MsgEmptyBody() string { return renderKey(KeyEmptyBody, "", "") }

// --- i18n renderer ---
// templates are sourced from pkg/i18n catalogs (YAML-driven)

func renderKey(key MsgKey, field, param string) string {
	if param != "" {
		return i18n.RenderLocale("en", string(key), field, param)
	}
	if field != "" {
		return i18n.RenderLocale("en", string(key), field)
	}
	return i18n.RenderLocale("en", string(key))
}

// renderKeyLocale uses provided locale.
func renderKeyLocale(locale string, key MsgKey, field, param string) string {
	if param != "" {
		return i18n.RenderLocale(locale, string(key), field, param)
	}
	if field != "" {
		return i18n.RenderLocale(locale, string(key), field)
	}
	return i18n.RenderLocale(locale, string(key))
}

// lookupTemplate removed: direct RenderLocale with parameters is used
