package security

import (
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct{ cost int }

// NewBcryptHasher creates a hasher with an injected cost. If cost is out of range (<=0),
// bcrypt.DefaultCost is used.
func NewBcryptHasher(cost int) *BcryptHasher { return &BcryptHasher{cost: cost} }

func (b *BcryptHasher) Hash(raw string) (string, error) {
	cost := b.cost
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (b *BcryptHasher) Compare(hashed string, raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw)) == nil
}
