package userusecase

import (
    "context"

    "appsechub/internal/application/dto"
    appsrv "appsechub/internal/application/service"
    "appsechub/internal/domain/user"
)

// UserUsecases defines the application-facing interface for user-related operations.
type UserUsecases interface {
    Register(ctx context.Context, input dto.CreateUserRequest) (*dto.UserResponse, error)
    Login(ctx context.Context, input dto.LoginRequest) (*dto.LoginResponse, error)
    GetMe(ctx context.Context, userID string) (*dto.UserResponse, error)
    ChangePassword(ctx context.Context, userID string, input dto.ChangePasswordRequest) error
}

// userUsecasesAggregator is a thin wrapper delegating to concrete use cases.
type userUsecasesAggregator struct {
    create *CreateUserUseCase
    login  *LoginUserUseCase
    getMe  *GetMeUseCase
    change *ChangePasswordUseCase
}

func NewUserUsecases(repo user.Repository, hasher PasswordHasher, jwt appsrv.TokenService) UserUsecases {
    return &userUsecasesAggregator{
        create: NewCreateUserUseCase(repo, hasher),
        login:  NewLoginUserUseCase(repo, hasher, jwt),
        getMe:  NewGetMeUseCase(repo),
        change: NewChangePasswordUseCase(repo, hasher),
    }
}

func (u *userUsecasesAggregator) Register(ctx context.Context, input dto.CreateUserRequest) (*dto.UserResponse, error) {
    return u.create.Execute(ctx, input)
}

func (u *userUsecasesAggregator) Login(ctx context.Context, input dto.LoginRequest) (*dto.LoginResponse, error) {
    return u.login.Execute(ctx, input)
}

func (u *userUsecasesAggregator) GetMe(ctx context.Context, userID string) (*dto.UserResponse, error) {
    return u.getMe.Execute(ctx, userID)
}

func (u *userUsecasesAggregator) ChangePassword(ctx context.Context, userID string, input dto.ChangePasswordRequest) error {
    return u.change.Execute(ctx, userID, input)
}

