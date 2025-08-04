package handler

import (
	"net/http"

	"appsechub/internal/application/dto"
	"appsechub/internal/application/usecase/userusecase"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	Usecase userusecase.UserUsecase
}

func NewUserHandler(uc userusecase.UserUsecase) *UserHandler {
    return &UserHandler{Usecase: uc}
}

func (h *UserHandler) Register(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    res, err := h.Usecase.CreateUser(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, res)
}

func (h *UserHandler) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    resp, err := h.Usecase.Login.Execute(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, resp)
}
