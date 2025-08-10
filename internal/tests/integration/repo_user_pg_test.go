//go:build integration

package integration

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	domuser "gostartkit/internal/domain/user"
	infdb "gostartkit/internal/infras/db"
	pgstore "gostartkit/internal/infras/storage/postgres"
)

func TestPostgres_UserRepository_CRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	host := getenvOr("DB_HOST", "localhost")
	port := getenvOr("DB_PORT", "5432")
	user := getenvOr("DB_USER", "gostartkit")
	pass := getenvOr("DB_PASSWORD", "devpassword")
	name := getenvOr("DB_NAME", "gostartkit")
	ssl := getenvOr("DB_SSLMODE", "disable")

	dsn := infdb.BuildPostgresDSN(host, port, user, pass, name, ssl)

	// Run migrations from repo root /migrations
	_, filename, _, _ := runtime.Caller(0)
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../.."))
	migrationsPath := filepath.Join(repoRoot, "migrations")
	infdb.RunMigrations(dsn, migrationsPath)

	db, err := pgstore.NewPostgresConnection(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := pgstore.NewUserRepository(db)

	// Create user
	email, _ := domuser.NewEmail("it@example.com")
	u := domuser.NewUser("It", "Test", email, "hashed:pass", domuser.RoleUser)
	if err := repo.Save(ctx, u); err != nil {
		t.Fatalf("save user: %v", err)
	}

	// Get by ID
	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Email != email {
		t.Fatalf("email mismatch: %v != %v", got.Email, email)
	}

	// Update
	got.FirstName = "It2"
	got.UpdatedAt = time.Now()
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Get by Email
	byEmail, err := repo.GetByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if byEmail.FirstName != "It2" {
		t.Fatalf("expected updated first name, got %s", byEmail.FirstName)
	}

	// List
	all, err := repo.GetAll(ctx)
	if err != nil || len(all) == 0 {
		t.Fatalf("get all: %v len=%d", err, len(all))
	}
}
