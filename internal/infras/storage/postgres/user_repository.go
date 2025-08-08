package postgres

import (
    "context"
    "database/sql"
    "errors"
    "time"

    domuser "appsechub/internal/domain/user"
    "github.com/google/uuid"
    pq "github.com/lib/pq"
)

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, u *domuser.User) error {
    cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    _, err := r.db.ExecContext(
        cctx,
        `INSERT INTO users (id, first_name, last_name, email, password, role, created_at, updated_at)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
        u.ID, u.FirstName, u.LastName, u.Email.String(), u.Password, string(u.Role), u.CreatedAt, u.UpdatedAt,
    )
    if err != nil {
        var pgErr *pq.Error
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
    row := r.db.QueryRowContext(cctx, `SELECT id, first_name, last_name, email, password, role, created_at, updated_at FROM users WHERE id=$1`, id)
    u, err := scanUser(row)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domuser.ErrUserNotFound
        }
        return nil, err
    }
    return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email domuser.Email) (*domuser.User, error) {
    cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    row := r.db.QueryRowContext(cctx, `SELECT id, first_name, last_name, email, password, role, created_at, updated_at FROM users WHERE email=$1`, email.String())
    u, err := scanUser(row)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domuser.ErrUserNotFound
        }
        return nil, err
    }
    return u, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*domuser.User, error) {
    cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    rows, err := r.db.QueryContext(cctx, `SELECT id, first_name, last_name, email, password, role, created_at, updated_at FROM users ORDER BY created_at DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var list []*domuser.User
    for rows.Next() {
        var u domuser.User
        var email string
        var role string
        if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &email, &u.Password, &role, &u.CreatedAt, &u.UpdatedAt); err != nil {
            return nil, err
        }
        u.Email = domuser.Email(email)
        u.Role = domuser.Role(role)
        list = append(list, &u)
    }
    return list, rows.Err()
}

func (r *UserRepository) Update(ctx context.Context, u *domuser.User) error {
    cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    _, err := r.db.ExecContext(
        cctx,
        `UPDATE users SET first_name=$2,last_name=$3,email=$4,password=$5,role=$6,updated_at=$7 WHERE id=$1`,
        u.ID, u.FirstName, u.LastName, u.Email.String(), u.Password, string(u.Role), u.UpdatedAt,
    )
    return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
    cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    _, err := r.db.ExecContext(cctx, `DELETE FROM users WHERE id=$1`, id)
    return err
}

func scanUser(row rowScanner) (*domuser.User, error) {
    var u domuser.User
    var email string
    var role string
    if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &email, &u.Password, &role, &u.CreatedAt, &u.UpdatedAt); err != nil {
        return nil, err
    }
    u.Email = domuser.Email(email)
    u.Role = domuser.Role(role)
    return &u, nil
}

type rowScanner interface {
    Scan(dest ...any) error
}

