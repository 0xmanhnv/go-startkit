package http

import (
	"gostartkit/internal/config"
	"gostartkit/internal/interfaces/http/handler"
	"gostartkit/internal/interfaces/http/router"

	"github.com/gin-gonic/gin"
)

// NewRouter constructs a Gin engine with config (prod-ready). Always require cfg.
// Deprecated: Implementation moved to internal/interfaces/http/router. This function delegates to router.New.
func NewRouter(userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) *gin.Engine {
	return router.New(userHandler, cfg, authMiddleware...)
}

// legacy file preserved to export NewRouter symbol only. Do not add logic here.
