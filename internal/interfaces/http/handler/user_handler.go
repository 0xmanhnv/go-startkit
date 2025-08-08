package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"appsechub/internal/application/dto"
	"appsechub/internal/application/usecase/userusecase"
	domuser "appsechub/internal/domain/user"
	"appsechub/internal/interfaces/http/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{ uc userusecase.UserUsecases }

func NewUserHandler(uc userusecase.UserUsecases) *UserHandler { return &UserHandler{uc: uc} }

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, msg := mapBindJSONError(err)
		response.BadRequest(c, code, msg)
		return
	}

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
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, msg := mapBindJSONError(err)
		response.BadRequest(c, code, msg)
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

// Refresh exchanges a valid refresh token for a new access token (and rotated refresh token if applicable).
func (h *UserHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		code, msg := mapBindJSONError(err)
		response.BadRequest(c, code, msg)
		return
	}
	resp, err := h.uc.Refresh(c.Request.Context(), body.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid_refresh_token", "refresh token invalid or expired")
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
		code, msg := mapBindJSONError(err)
		response.BadRequest(c, code, msg)
		return
	}
	if err := h.uc.Logout(c.Request.Context(), body.RefreshToken); err != nil {
		response.Unauthorized(c, "invalid_refresh_token", "refresh token invalid or expired")
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
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		code, msg := mapBindJSONError(err)
		response.BadRequest(c, code, msg)
		return
	}
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

// mapBindJSONError chuyển lỗi bind JSON của Gin thành thông điệp thân thiện với người dùng.
func mapBindJSONError(err error) (code, message string) {
	// Trường hợp body rỗng
	if errors.Is(err, io.EOF) {
		return "invalid_request", "request body is empty"
	}
	// JSON không hợp lệ (syntax)
	var syn *json.SyntaxError
	if errors.As(err, &syn) {
		return "invalid_request", fmt.Sprintf("malformed JSON at position %d", syn.Offset)
	}
	// Sai kiểu dữ liệu cho field cụ thể
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		if typeErr.Field != "" {
			return "invalid_request", fmt.Sprintf("invalid type for field %s", typeErr.Field)
		}
		return "invalid_request", "invalid type in JSON payload"
	}
	// Mặc định
	return "invalid_request", "invalid JSON payload"
}
