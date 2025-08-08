package handler

import (
    "appsechub/internal/application/dto"
    "appsechub/internal/application/usecase/userusecase"
    "appsechub/internal/interfaces/http/response"
    domuser "appsechub/internal/domain/user"

    "github.com/gin-gonic/gin"
)

type UserHandler struct{ uc userusecase.UserUsecases }

func NewUserHandler(uc userusecase.UserUsecases) *UserHandler { return &UserHandler{uc: uc} }

func (h *UserHandler) Register(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "invalid_request", err.Error())
        return
    }

    res, err := h.uc.Register(c.Request.Context(), req)
    if err != nil {
        switch err {
        case domuser.ErrEmailAlreadyExists:
            response.BadRequest(c, "email_exists", "email already exists")
        case domuser.ErrInvalidRole, domuser.ErrInvalidEmail:
            response.BadRequest(c, "invalid_request", err.Error())
        default:
            response.InternalError(c, "server_error", "internal error")
        }
        return
    }
    response.Created(c, res)
}

func (h *UserHandler) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "invalid_request", err.Error())
        return
    }
    resp, err := h.uc.Login(c.Request.Context(), req)
    if err != nil {
        switch err {
        case domuser.ErrUserNotFound:
            response.Unauthorized(c, "invalid_credentials", "email or password is incorrect")
        default:
            response.Unauthorized(c, "invalid_credentials", "email or password is incorrect")
        }
        return
    }
    response.OK(c, resp)
}

func (h *UserHandler) GetMe(c *gin.Context) {
    userID := c.GetString("user_id")
    if userID == "" {
        response.Unauthorized(c, "unauthorized", "missing user context")
        return
    }
    res, err := h.uc.GetMe(c.Request.Context(), userID)
    if err != nil {
        switch err {
        case domuser.ErrUserNotFound:
            response.BadRequest(c, "not_found", "user not found")
        default:
            response.InternalError(c, "server_error", "internal error")
        }
        return
    }
    response.OK(c, res)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
    userID := c.GetString("user_id")
    if userID == "" {
        response.Unauthorized(c, "unauthorized", "missing user context")
        return
    }
    var req dto.ChangePasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "invalid_request", err.Error())
        return
    }
    if err := h.uc.ChangePassword(c.Request.Context(), userID, req); err != nil {
        switch err {
        case domuser.ErrInvalidPassword:
            response.BadRequest(c, "invalid_current_password", "current password is incorrect")
        case domuser.ErrUserNotFound:
            response.BadRequest(c, "not_found", "user not found")
        default:
            response.InternalError(c, "server_error", "internal error")
        }
        return
    }
    response.OK(c, gin.H{"changed": true})
}
