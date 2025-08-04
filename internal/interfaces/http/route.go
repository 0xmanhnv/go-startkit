package http

import (
	"appsechub/internal/interfaces/http/handler"
	"github.com/gin-gonic/gin"
)

func NewRouter(userHandler *handler.UserHandler) *gin.Engine {
	r := gin.Default()
	
	// Routes
	v1 := r.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
		}
	}

	return r
}
