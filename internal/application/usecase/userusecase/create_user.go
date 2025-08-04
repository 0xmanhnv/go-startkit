package userusecase

import (
	"context"

	"appsechub/internal/domain/user"
)

type CreateUserUseCase struct {
    repo user.Repository
    hasher PasswordHasher
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input dto.CreateUserRequest) (*dto.UserResponse, error) {
    hashed := uc.hasher.Hash(input.Password)
    user := user.NewUser(input.Email, hashed)
    err := uc.repo.Save(ctx, user)
    return &dto.UserResponse{ID: user.ID, Email: user.Email}, err
}