package userusecase

import (
	"context"

	"gostartkit/internal/application/dto"
	domuser "gostartkit/internal/domain/user"

	"github.com/google/uuid"
)

type ChangePasswordUseCase struct {
	repo   domuser.Repository
	hasher PasswordHasher
}

func NewChangePasswordUseCase(repo domuser.Repository, hasher PasswordHasher) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{repo: repo, hasher: hasher}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID string, input dto.ChangePasswordRequest) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return domuser.ErrInvalidID
	}
	u, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !uc.hasher.Compare(u.Password, input.CurrentPassword) {
		return domuser.ErrInvalidPassword
	}
	hashed, err := uc.hasher.Hash(input.NewPassword)
	if err != nil {
		return err
	}
	u.Password = hashed
	return uc.repo.Update(ctx, u)
}
