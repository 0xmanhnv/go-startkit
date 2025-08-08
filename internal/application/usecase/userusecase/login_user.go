package userusecase

import (
    "context"
    "errors"

    "appsechub/internal/application/dto"
    appsrv "appsechub/internal/application/service"
    "appsechub/internal/domain/user"
)

type LoginUserUseCase struct {
    repo   user.Repository
    hasher PasswordHasher
    jwt    appsrv.TokenService
}

func (uc *LoginUserUseCase) Execute(ctx context.Context, input dto.LoginRequest) (*dto.LoginResponse, error) {
    emailVO, err := user.NewEmail(input.Email)
    if err != nil {
        return nil, err
    }

    u, err := uc.repo.GetByEmail(ctx, emailVO)
    if err != nil {
        return nil, errors.New("invalid credentials")
    }
    if !uc.hasher.Compare(u.Password, input.Password) {
        return nil, errors.New("invalid credentials")
    }

    token, err := uc.jwt.GenerateToken(u.ID.String(), string(u.Role))
    if err != nil {
        return nil, err
    }

    return &dto.LoginResponse{
        AccessToken: token,
        User: dto.UserResponse{
            ID:        u.ID,
            Email:     u.Email.String(),
            FirstName: u.FirstName,
            LastName:  u.LastName,
            CreatedAt: u.CreatedAt,
        },
    }, nil
}

