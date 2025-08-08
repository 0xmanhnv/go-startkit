package userusecase

import (
    appsrv "appsechub/internal/application/service"
    "appsechub/internal/domain/user"
)

func NewCreateUserUseCase(repo user.Repository, hasher PasswordHasher) *CreateUserUseCase {
    return &CreateUserUseCase{repo: repo, hasher: hasher}
}

func NewLoginUserUseCase(repo user.Repository, hasher PasswordHasher, jwt appsrv.TokenService) *LoginUserUseCase {
    return &LoginUserUseCase{repo: repo, hasher: hasher, jwt: jwt}
}

