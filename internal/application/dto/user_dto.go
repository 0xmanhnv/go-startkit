package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
    Email     string    `json:"email"`
    FirstName string    `json:"first_name"`
    LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
    FirstName string `json:"first_name" binding:"required,min=1"`
    LastName  string `json:"last_name" binding:"required,min=1"`
    Email     string `json:"email" binding:"required,email"`
    Password  string `json:"password" binding:"required,min=8"`
    Role      string `json:"role" binding:"required"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    AccessToken string       `json:"access_token"`
    User        UserResponse `json:"user"`
}

type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" binding:"required,min=8"`
    NewPassword     string `json:"new_password" binding:"required,min=8"`
}
