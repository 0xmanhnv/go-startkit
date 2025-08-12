package userusecase

// PasswordHasher defines hashing interface used by use cases
type PasswordHasher interface {
	// Hash returns the hashed representation of the raw password or an error.
	Hash(raw string) (string, error)
	Compare(hashed string, raw string) bool
}
