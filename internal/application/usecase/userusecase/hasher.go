package userusecase

// PasswordHasher defines hashing interface used by use cases
type PasswordHasher interface {
    Hash(raw string) string
    Compare(hashed string, raw string) bool
}

