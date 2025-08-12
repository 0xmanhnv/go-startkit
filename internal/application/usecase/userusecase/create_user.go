package userusecase

import (
	"context"

	"gostartkit/internal/application/dto"
	"gostartkit/internal/domain/user"
)

type CreateUserUseCase struct {
	repo   user.Repository
	hasher PasswordHasher
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input dto.CreateUserRequest) (*dto.UserResponse, error) {
	emailVO, err := user.NewEmail(input.Email)
	if err != nil {
		return nil, err
	}
	role := user.Role(input.Role)
	if !role.IsValid() {
		return nil, user.ErrInvalidRole
	}
	// Enforce public registration only for non-admin users
	if role == user.RoleAdmin {
		return nil, user.ErrInvalidRole
	}
	hashed, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	newUser := user.NewUser(input.FirstName, input.LastName, emailVO, hashed, role)
	if err := user.ValidateUser(newUser); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, newUser); err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		ID:        newUser.ID,
		Email:     newUser.Email.String(),
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		CreatedAt: newUser.CreatedAt,
	}, nil
}
