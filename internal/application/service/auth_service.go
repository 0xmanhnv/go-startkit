package authservice

import (
	"context"

	"appsechub/internal/domain"
)

type AuthService interface {
    HashPassword(raw string) string
    ComparePassword(hashed string, raw string) bool
    GenerateToken(userID, role string) (string, error)
}
