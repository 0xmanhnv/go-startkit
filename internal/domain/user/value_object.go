package user

import (
	appval "gostartkit/pkg/validator"
	"strings"
)

// Email
type Email string

func (e Email) String() string {
	return string(e)
}

func (e Email) IsValid() bool {
	return appval.IsValidEmail(string(e))
}

func NewEmail(s string) (Email, error) {
	email := Email(strings.TrimSpace(s))
	if !email.IsValid() {
		return "", ErrInvalidEmail
	}
	return email, nil
}

// Password
type Password struct {
	hash string
}

func NewPassword(hash string) Password {
	return Password{hash: hash}
}

func (p Password) String() string {
	return p.hash
}

// Role
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser, RoleViewer:
		return true
	default:
		return false
	}
}
