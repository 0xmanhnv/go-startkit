package main

import (
	"context"
	"time"

	"gostartkit/internal/application/usecase/userusecase"
	"gostartkit/internal/config"
	domuser "gostartkit/internal/domain/user"
	pgstore "gostartkit/internal/infras/storage/postgres"
	"gostartkit/pkg/logger"
)

// seedInitialUser ensures an initial user exists using values from config.Seed.
func seedInitialUser(_ any, repo *pgstore.UserRepository, hasher userusecase.PasswordHasher, cfg *config.Config) error {
	// Validate inputs
	if cfg.Seed.Email == "" || cfg.Seed.Password == "" {
		logger.L().Warn("seed_skipped_missing_env")
		return nil
	}
	emailVO, err := domuser.NewEmail(cfg.Seed.Email)
	if err != nil {
		return err
	}
	// Check exists
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := repo.GetByEmail(ctx, emailVO); err == nil {
		// exists
		return nil
	}
	// Create
	role := domuser.Role(cfg.Seed.Role)
	if !role.IsValid() {
		role = domuser.RoleAdmin
	}
	hashed, err := hasher.Hash(cfg.Seed.Password)
	if err != nil {
		return err
	}
	u := domuser.NewUser(cfg.Seed.FirstName, cfg.Seed.LastName, emailVO, hashed, role)
	if err := repo.Save(ctx, u); err != nil {
		return err
	}
	logger.L().Info("seed_success", "email", u.Email.String(), "role", role)
	return nil
}
