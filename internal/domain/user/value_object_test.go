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
	_, err := NewEmail("invalid-email")
	if err == nil {
		t.Fatalf("expected error for invalid email")
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
