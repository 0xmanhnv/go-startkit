package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"
	"appsechub/internal/application/dto"
	"appsechub/internal/application/usecase/userusecase"

	"github.com/gin-gonic/gin"
)

type fakeUserUC struct{}

func (fakeUserUC) Register(_ context.Context, _ dto.CreateUserRequest) (*dto.UserResponse, error) {
	return nil, nil
}
func (fakeUserUC) Login(_ context.Context, input dto.LoginRequest) (*dto.LoginResponse, error) {
	return &dto.LoginResponse{AccessToken: "token", User: dto.UserResponse{Email: input.Email}}, nil
}
func (fakeUserUC) GetMe(_ context.Context, _ string) (*dto.UserResponse, error) { return nil, nil }
func (fakeUserUC) ChangePassword(_ context.Context, _ string, _ dto.ChangePasswordRequest) error {
	return nil
}
func (fakeUserUC) Refresh(_ context.Context, _ string) (*dto.LoginResponse, error) { return nil, nil }
func (fakeUserUC) Logout(_ context.Context, _ string) error                        { return nil }

var _ userusecase.UserUsecases = (*fakeUserUC)(nil)

func TestUserHandler_Login_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := fakeUserUC{}
	h := NewUserHandler(uc)
	r := gin.New()
	r.POST("/v1/auth/login", func(c *gin.Context) {
		// simulate validate middleware placing DTO into context
		var req dto.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Set("req", req)
		h.Login(c)
	})

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "pass"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
