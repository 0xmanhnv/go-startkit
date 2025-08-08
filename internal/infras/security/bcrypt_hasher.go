package security

import (
    "golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher { return &BcryptHasher{} }

func (b *BcryptHasher) Hash(raw string) string {
    hashed, _ := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
    return string(hashed)
}

func (b *BcryptHasher) Compare(hashed string, raw string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw)) == nil
}

