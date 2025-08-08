package service

// TokenService abstracts token generation for application layer
type TokenService interface {
    GenerateToken(userID string, role string) (string, error)
}

