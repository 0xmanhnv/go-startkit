package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"
	"gostartkit/internal/application/dto"
	"gostartkit/internal/interfaces/http/handler"

	"github.com/gin-gonic/gin"
)

type ucStub struct{}

func (ucStub) Register(context.Context, dto.CreateUserRequest) (*dto.UserResponse, error) {
	return nil, nil
}
func (ucStub) Login(context.Context, dto.LoginRequest) (*dto.LoginResponse, error) {
	return &dto.LoginResponse{AccessToken: "token", User: dto.UserResponse{Email: "user@example.com"}}, nil
}
func (ucStub) GetMe(context.Context, string) (*dto.UserResponse, error)                { return nil, nil }
func (ucStub) ChangePassword(context.Context, string, dto.ChangePasswordRequest) error { return nil }
func (ucStub) Refresh(context.Context, string) (*dto.LoginResponse, error)             { return nil, nil }
func (ucStub) Logout(context.Context, string) error                                    { return nil }

func TestLogin_Handler_ExternalPackage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewUserHandler(ucStub{})
	r := gin.New()
	r.POST("/v1/auth/login", func(c *gin.Context) {
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Set("req", req)
		h.Login(c)
	})
	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "x"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
