package http

import (
	"context"
	"net/http"
	"time"

	"appsechub/internal/config"
	"appsechub/internal/interfaces/http/handler"

	"github.com/gin-gonic/gin"
)

// NewRouterWithReadiness allows injecting a DB ping function for a real readiness check.
func NewRouterWithReadiness(userHandler *handler.UserHandler, cfg *config.Config, ping func(ctx context.Context) error) *gin.Engine {
	r := NewRouter(userHandler, cfg)
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		if err := ping(ctx); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})
	return r
}

// AddReadiness registers /readyz on an existing router with a provided ping function
func AddReadiness(r *gin.Engine, ping func(ctx context.Context) error) {
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()
		if err := ping(ctx); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})
}
