package response

import (
	"errors"
	"testing"

	"gostartkit/internal/application/apperr"
	domuser "gostartkit/internal/domain/user"
)

func TestFromError_Mapping(t *testing.T) {
	cases := []struct {
		in         error
		wantStatus int
	}{
		{apperr.ErrInvalidCredentials, 401},
		{apperr.ErrInvalidRefreshToken, 401},
		{domuser.ErrUserNotFound, 404},
		{domuser.ErrEmailAlreadyExists, 409},
		{errors.New("x"), 500},
	}
	for _, c := range cases {
		got, _, _ := FromError(c.in)
		if got != c.wantStatus {
			t.Fatalf("status = %d, want %d", got, c.wantStatus)
		}
	}
}
