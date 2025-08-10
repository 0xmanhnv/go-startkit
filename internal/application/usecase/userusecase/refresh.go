package userusecase

import (
	"context"
	"errors"

	"gostartkit/internal/application/apperr"
	"gostartkit/internal/application/dto"
	"gostartkit/internal/application/ports"
	"gostartkit/internal/domain/user"

	"github.com/google/uuid"
)

// RefreshUseCase handles refresh token rotation and access token issuance.
type RefreshUseCase struct {
	repo  user.Repository
	jwt   ports.TokenIssuer
	store ports.RefreshTokenStore
	// refreshTTLSeconds controls how long newly issued refresh tokens live.
	// If non-positive, a sensible default will be used.
	refreshTTLSeconds int
}

func NewRefreshUseCase(repo user.Repository, jwt ports.TokenIssuer) *RefreshUseCase {
	return &RefreshUseCase{repo: repo, jwt: jwt}
}

func NewRefreshUseCaseWithStore(repo user.Repository, jwt ports.TokenIssuer, store ports.RefreshTokenStore, refreshTTLSeconds int) *RefreshUseCase {
	return &RefreshUseCase{repo: repo, jwt: jwt, store: store, refreshTTLSeconds: refreshTTLSeconds}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	if uc.store == nil {
		return nil, apperr.ErrRefreshStoreNotConfigured
	}
	ttl := uc.refreshTTLSeconds
	if ttl <= 0 {
		ttl = 3600 * 24 * 7 // default 7 days
	}
	newRefresh, userID, err := uc.store.Rotate(ctx, refreshToken, ttl)
	if err != nil {
		return nil, apperr.ErrInvalidRefreshToken
	}
	u, err := uc.repo.GetByID(ctx, mustParseUUID(userID))
	if err != nil {
		return nil, apperr.ErrInvalidRefreshToken
	}
	access, err := uc.jwt.GenerateToken(u.ID.String(), string(u.Role))
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{AccessToken: access, RefreshToken: newRefresh, User: dto.UserResponse{ID: u.ID, Email: u.Email.String(), FirstName: u.FirstName, LastName: u.LastName, CreatedAt: u.CreatedAt}}, nil
}

func (uc *RefreshUseCase) Revoke(ctx context.Context, refreshToken string) error {
	if uc.store == nil {
		return errors.New("refresh store not configured")
	}
	return uc.store.Revoke(ctx, refreshToken)
}

// mustParseUUID is a tiny helper; in real code prefer explicit error handling.
func mustParseUUID(s string) uuid.UUID { id, _ := uuid.Parse(s); return id }
