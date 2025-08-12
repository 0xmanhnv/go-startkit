package validator

import "testing"

func TestIsValidEmail_Valid(t *testing.T) {
	cases := []string{
		"user@example.com",
		"u.x+tag@sub.example.co.uk",
		"USER@EXAMPLE.ORG",
		"user_name-123@example.io",
		" user@example.com ", // trimming allowed
	}
	for _, in := range cases {
		if !IsValidEmail(in) {
			t.Fatalf("expected valid email: %q", in)
		}
	}
}

func TestIsValidEmail_Invalid(t *testing.T) {
	cases := []string{
		"",
		"invalid",
		"user@",
		"@example.com",
		"User <user@example.com>", // display name not allowed
		"user@exa mple.com",       // spaces not allowed
		"user@example..com",       // consecutive dots (ParseAddress may accept but often invalid),
	}
	for _, in := range cases {
		if IsValidEmail(in) {
			t.Fatalf("expected invalid email: %q", in)
		}
	}
}

func TestIsStrongPassword(t *testing.T) {
	valid := []string{
		"Aa1!aaaaaaaa",
		"Str0ng#Passw0rd",
		"P@ssw0rdIsLongEnough",
	}
	for _, s := range valid {
		if !IsStrongPassword(s) {
			t.Fatalf("expected strong password: %q", s)
		}
	}
	invalid := []string{
		"short1!A",
		"nouppercase1!aaaaaaaa",
		"NOLOWERCASE1!AAAAAAAA",
		"NoDigits!!!!!!!!",
		"NoSpecials1111111",
		"Has space 123!Aa",
		"aaaaaaaaaaaa",     // all same
		"Password123!abcd", // contains common word
	}
	for _, s := range invalid {
		if IsStrongPassword(s) {
			t.Fatalf("expected weak password: %q", s)
		}
	}
}
