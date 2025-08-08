package userusecase

import (
	"context"
	"errors"

	"appsechub/internal/application/dto"
	"appsechub/internal/application/ports"
	"appsechub/internal/domain/user"
)

type LoginUserUseCase struct {
	repo   user.Repository
	hasher PasswordHasher
	jwt    ports.TokenIssuer
	store  ports.RefreshTokenStore
}

func (uc *LoginUserUseCase) Execute(ctx context.Context, input dto.LoginRequest) (*dto.LoginResponse, error) {
	emailVO, err := user.NewEmail(input.Email)
	if err != nil {
		return nil, err
	}

	u, err := uc.repo.GetByEmail(ctx, emailVO)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if !uc.hasher.Compare(u.Password, input.Password) {
		return nil, errors.New("invalid credentials")
	}

	token, err := uc.jwt.GenerateToken(u.ID.String(), string(u.Role))
	if err != nil {
		return nil, err
	}
	var refresh string
	if uc.store != nil {
		refresh, _ = uc.store.Issue(ctx, u.ID.String(), 3600*24*7) // 7d default
	}

	return &dto.LoginResponse{
		AccessToken:  token,
		RefreshToken: refresh,
		User: dto.UserResponse{
			ID:        u.ID,
			Email:     u.Email.String(),
			FirstName: u.FirstName,
			LastName:  u.LastName,
			CreatedAt: u.CreatedAt,
		},
	}, nil
}
