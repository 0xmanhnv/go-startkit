package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	GetAll(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
