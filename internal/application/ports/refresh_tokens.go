package ports

import "context"

// RefreshTokenStore abstracts storing and rotating refresh tokens per user/session.
type RefreshTokenStore interface {
	// Issue creates a new refresh token for a user and returns the token string.
	Issue(ctx context.Context, userID string, ttlSeconds int) (string, error)
	// Rotate invalidates the old token and issues a new one atomically.
	Rotate(ctx context.Context, oldToken string, ttlSeconds int) (newToken string, userID string, err error)
	// Revoke invalidates a specific token.
	Revoke(ctx context.Context, token string) error
	// Validate returns the userID if the token is valid (not revoked/expired).
	Validate(ctx context.Context, token string) (userID string, err error)
}
