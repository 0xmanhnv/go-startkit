package main

import (
    "context"
    "log"
    "time"

    "appsechub/internal/config"
    domuser "appsechub/internal/domain/user"
    "appsechub/internal/application/usecase/userusecase"
    pgstore "appsechub/internal/infras/storage/postgres"
)

// seedInitialUser ensures an initial user exists using values from config.Seed.
func seedInitialUser(sqlDB any, repo *pgstore.UserRepository, hasher userusecase.PasswordHasher, cfg *config.Config) error {
    // Validate inputs
    if cfg.Seed.Email == "" || cfg.Seed.Password == "" {
        log.Printf("seed skipped: SEED_USER_EMAIL or SEED_USER_PASSWORD is empty")
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
    hashed := hasher.Hash(cfg.Seed.Password)
    u := domuser.NewUser(cfg.Seed.FirstName, cfg.Seed.LastName, emailVO, hashed, role)
    if err := repo.Save(ctx, u); err != nil {
        return err
    }
    log.Printf("seeded initial user: %s (%s)", u.Email.String(), role)
    return nil
}

