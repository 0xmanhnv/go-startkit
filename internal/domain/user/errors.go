package user

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUserRole    = errors.New("invalid user role")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidFirstName   = errors.New("invalid first name")
	ErrInvalidLastName    = errors.New("invalid last name")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidID          = errors.New("invalid ID")
)
