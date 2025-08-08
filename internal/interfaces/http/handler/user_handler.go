package handler

import (
	"appsechub/internal/application/dto"
	"appsechub/internal/application/usecase/userusecase"
	domuser "appsechub/internal/domain/user"
	"appsechub/internal/interfaces/http/response"
	"appsechub/internal/interfaces/http/validation"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{ uc userusecase.UserUsecases }

func NewUserHandler(uc userusecase.UserUsecases) *UserHandler { return &UserHandler{uc: uc} }

func (h *UserHandler) Register(c *gin.Context) {
	req := c.MustGet("req").(dto.CreateUserRequest)

	res, err := h.uc.Register(c.Request.Context(), req)
	if err != nil {
		switch err {
		case domuser.ErrEmailAlreadyExists:
			response.Conflict(c, "email_exists", "email already exists")
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
	req := c.MustGet("req").(dto.LoginRequest)
	resp, err := h.uc.Login(c.Request.Context(), req)
	if err != nil {
		status, code, msg := response.FromError(err)
		switch status {
		case 400:
			response.BadRequest(c, code, msg)
		case 401:
			response.Unauthorized(c, code, msg)
		case 404:
			response.NotFound(c, code, msg)
		case 409:
			response.Conflict(c, code, msg)
		default:
			response.InternalError(c, code, msg)
		}
		return
	}
	response.OK(c, resp)
}

// Refresh exchanges a valid refresh token for a new access token (and rotated refresh token if applicable).
func (h *UserHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		code, msg := validation.MapBindJSONError(err)
		response.BadRequest(c, code, msg)
		return
	}
	resp, err := h.uc.Refresh(c.Request.Context(), body.RefreshToken)
	if err != nil {
		status, code, msg := response.FromError(err)
		switch status {
		case 400:
			response.BadRequest(c, code, msg)
		case 401:
			response.Unauthorized(c, code, msg)
		case 404:
			response.NotFound(c, code, msg)
		case 409:
			response.Conflict(c, code, msg)
		default:
			response.InternalError(c, code, msg)
		}
		return
	}
	response.OK(c, resp)
}

// Logout revokes a refresh token (log out current session).
func (h *UserHandler) Logout(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		code, msg := validation.MapBindJSONError(err)
		response.BadRequest(c, code, msg)
		return
	}
	if err := h.uc.Logout(c.Request.Context(), body.RefreshToken); err != nil {
		status, code, msg := response.FromError(err)
		switch status {
		case 400:
			response.BadRequest(c, code, msg)
		case 401:
			response.Unauthorized(c, code, msg)
		case 404:
			response.NotFound(c, code, msg)
		case 409:
			response.Conflict(c, code, msg)
		default:
			response.InternalError(c, code, msg)
		}
		return
	}
	response.OK(c, gin.H{"revoked": true})
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
			response.NotFound(c, "not_found", "user not found")
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
	req := c.MustGet("req").(dto.ChangePasswordRequest)
	if err := h.uc.ChangePassword(c.Request.Context(), userID, req); err != nil {
		switch err {
		case domuser.ErrInvalidPassword:
			response.BadRequest(c, "invalid_current_password", "current password is incorrect")
		case domuser.ErrUserNotFound:
			response.NotFound(c, "not_found", "user not found")
		default:
			response.InternalError(c, "server_error", "internal error")
		}
		return
	}
	response.OK(c, gin.H{"changed": true})
}
