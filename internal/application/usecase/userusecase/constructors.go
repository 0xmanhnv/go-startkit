package userusecase

import (
	"gostartkit/internal/application/ports"
	"gostartkit/internal/domain/user"
)

func NewCreateUserUseCase(repo user.Repository, hasher PasswordHasher) *CreateUserUseCase {
	return &CreateUserUseCase{repo: repo, hasher: hasher}
}

func NewLoginUserUseCase(repo user.Repository, hasher PasswordHasher, jwt ports.TokenIssuer, store ports.RefreshTokenStore) *LoginUserUseCase {
	return &LoginUserUseCase{repo: repo, hasher: hasher, jwt: jwt, store: store}
}
