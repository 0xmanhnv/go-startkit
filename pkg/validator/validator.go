package validator

import (
	"net/mail"
	"strings"
)

// IsValidEmail validates an email address using net/mail and strict format.
// It accepts only plain addresses (no display name) and trims surrounding spaces.
func IsValidEmail(input string) bool {
	s := strings.TrimSpace(input)
	if s == "" {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return false
	}
	// Disallow display names; require exact address match
	if addr.Name != "" {
		return false
	}
	return addr.Address == s
}

// IsStrongPassword checks a password against baseline rules:
// - length >= 12
// - contains at least one lowercase, one uppercase, one digit, one special
// - no spaces or control characters
func IsStrongPassword(pw string) bool {
	s := strings.TrimSpace(pw)
	if len(s) < 12 {
		return false
	}
	// Quick reject: avoid trivial repetitions
	if allSameRune(s) {
		return false
	}
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		case r == ' ' || r < 32:
			return false
		default:
			hasSpecial = true
		}
	}
	if !(hasLower && hasUpper && hasDigit && hasSpecial) {
		return false
	}
	// Deny common weak patterns/words
	lower := strings.ToLower(s)
	if containsCommonPasswordSubstring(lower) {
		return false
	}
	return true
}

func allSameRune(s string) bool {
	if len(s) == 0 {
		return true
	}
	first := rune(s[0])
	for _, r := range s {
		if r != first {
			return false
		}
	}
	return true
}

func containsCommonPasswordSubstring(lower string) bool {
	common := []string{
		"password", "123456", "qwerty", "letmein", "welcome",
		"admin", "iloveyou", "abc123", "1q2w3e", "monkey",
	}
	for _, w := range common {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}
