package user

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     Email
	Password  string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(firstName, lastName string, email Email, password string, role Role) *User {
    // Domain invariants enforced at construction (basic checks)
    // Note: deeper validation (e.g., password strength) belongs to application layer
    u := &User{
        ID:        uuid.New(),
        FirstName: firstName,
        LastName:  lastName,
        Email:     email,
        Password:  password,
        Role:      role,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    return u
}