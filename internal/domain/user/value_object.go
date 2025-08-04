package user

import (
	"fmt"
	"strings"
)

// Email
type Email string

func (e Email) String() string {
    return string(e)
}

func (e Email) IsValid() bool {
    return strings.Contains(string(e), "@")
}

func NewEmail(s string) (Email, error) {
    email := Email(strings.TrimSpace(s))
    if !email.IsValid() {
        return "", fmt.Errorf("invalid email format")
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
