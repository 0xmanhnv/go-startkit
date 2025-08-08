package userusecase

import (
	"context"
	"errors"

	"appsechub/internal/application/dto"
	"appsechub/internal/application/ports"
	"appsechub/internal/domain/user"

	"github.com/google/uuid"
)

// RefreshUseCase handles refresh token rotation and access token issuance.
type RefreshUseCase struct {
	repo  user.Repository
	jwt   ports.TokenIssuer
	store ports.RefreshTokenStore
}

func NewRefreshUseCase(repo user.Repository, jwt ports.TokenIssuer) *RefreshUseCase {
	return &RefreshUseCase{repo: repo, jwt: jwt}
}

func NewRefreshUseCaseWithStore(repo user.Repository, jwt ports.TokenIssuer, store ports.RefreshTokenStore) *RefreshUseCase {
	return &RefreshUseCase{repo: repo, jwt: jwt, store: store}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	if uc.store == nil {
		return nil, errors.New("refresh store not configured")
	}
	newRefresh, userID, err := uc.store.Rotate(ctx, refreshToken, 3600*24*7) // 7 days default
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	u, err := uc.repo.GetByID(ctx, mustParseUUID(userID))
	if err != nil {
		return nil, errors.New("invalid refresh token")
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
