package user

import "testing"

func TestNewEmail_Valid(t *testing.T) {
	e, err := NewEmail("user@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if e.String() != "user@example.com" {
		t.Fatalf("unexpected email: %s", e)
	}
}

func TestNewEmail_Invalid(t *testing.T) {
	cases := []string{
		"invalid-email",
		"user@",
		"@example.com",
		"User <user@example.com>",
		" user@example.com ", // leading/trailing spaces should be trimmed; still valid
	}
	// The last case should be valid after trimming; test separately
	for _, in := range cases[:4] {
		if _, err := NewEmail(in); err == nil {
			t.Fatalf("expected error for invalid email: %q", in)
		}
	}
	// Valid after trimming
	if e, err := NewEmail(cases[4]); err != nil || e.String() != "user@example.com" {
		t.Fatalf("expected trimmed valid email, got %v %q", err, e)
	}
}

func TestRole_IsValid(t *testing.T) {
	cases := []struct {
		in Role
		ok bool
	}{
		{RoleAdmin, true}, {RoleUser, true}, {RoleViewer, true}, {Role("x"), false},
	}
	for _, c := range cases {
		if got := c.in.IsValid(); got != c.ok {
			t.Fatalf("role %q valid=%v, want %v", c.in, got, c.ok)
		}
	}
}
