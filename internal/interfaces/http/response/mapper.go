package response

import (
	"errors"

	"appsechub/internal/application/apperr"
	domuser "appsechub/internal/domain/user"
)

// FromError maps a domain/application error to HTTP status, code and safe message.
func FromError(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, apperr.ErrInvalidCredentials):
		return 401, CodeInvalidCredentials, MsgInvalidCredentials
	case errors.Is(err, apperr.ErrInvalidRefreshToken):
		return 401, CodeInvalidRefreshToken, MsgInvalidRefreshToken
	case errors.Is(err, domuser.ErrUserNotFound):
		return 404, CodeNotFound, MsgNotFound
	case errors.Is(err, domuser.ErrEmailAlreadyExists):
		return 409, CodeConflict, "email already exists"
	case errors.Is(err, domuser.ErrInvalidEmail), errors.Is(err, domuser.ErrInvalidRole):
		return 400, CodeInvalidRequest, "invalid request"
	default:
		return 500, CodeServerError, MsgServerError
	}
}
