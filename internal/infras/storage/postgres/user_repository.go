package postgres

import (
	"context"
	"errors"
	"time"

	domuser "appsechub/internal/domain/user"
	pstore "appsechub/internal/infras/storage/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
	q    *pstore.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool, q: pstore.New(pool)}
}

func (r *UserRepository) Save(ctx context.Context, u *domuser.User) error {
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := r.q.CreateUser(cctx, pstore.CreateUserParams{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email.String(),
		Password:  u.Password,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return domuser.ErrEmailAlreadyExists
			}
		}
		return err
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domuser.User, error) {
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	row, err := r.q.GetUserByID(cctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domuser.ErrUserNotFound
		}
		return nil, err
	}
	return &domuser.User{
		ID:        row.ID,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Email:     domuser.Email(row.Email),
		Password:  row.Password,
		Role:      domuser.Role(row.Role),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email domuser.Email) (*domuser.User, error) {
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	row, err := r.q.GetUserByEmail(cctx, email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domuser.ErrUserNotFound
		}
		return nil, err
	}
	return &domuser.User{
		ID:        row.ID,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Email:     domuser.Email(row.Email),
		Password:  row.Password,
		Role:      domuser.Role(row.Role),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*domuser.User, error) {
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := r.q.ListUsers(cctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domuser.User, 0, len(rows))
	for _, row := range rows {
		u := &domuser.User{
			ID:        row.ID,
			FirstName: row.FirstName,
			LastName:  row.LastName,
			Email:     domuser.Email(row.Email),
			Password:  row.Password,
			Role:      domuser.Role(row.Role),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
		out = append(out, u)
	}
	return out, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domuser.User) error {
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return r.q.UpdateUser(cctx, pstore.UpdateUserParams{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email.String(),
		Password:  u.Password,
		Role:      string(u.Role),
		UpdatedAt: u.UpdatedAt, // kept for explicitness; DB trigger also updates this
	})
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return r.q.DeleteUser(cctx, id)
}
