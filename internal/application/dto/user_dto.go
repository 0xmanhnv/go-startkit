package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    AccessToken string       `json:"access_token"`
    User        UserResponse `json:"user"`
}
