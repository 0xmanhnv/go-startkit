package ports

// TokenIssuer abstracts token issuance for application layer
type TokenIssuer interface {
	GenerateToken(userID string, role string) (string, error)
}
