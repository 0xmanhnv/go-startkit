package apperr

import "errors"

// Sentinel application-level errors for consistent handling/mapping.
var (
	ErrInvalidCredentials        = errors.New("invalid_credentials")
	ErrInvalidRefreshToken       = errors.New("invalid_refresh_token")
	ErrRefreshStoreNotConfigured = errors.New("refresh_store_not_configured")
)
