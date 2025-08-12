package router

import (
	"gostartkit/internal/config"
	"gostartkit/internal/interfaces/http/handler"
	"gostartkit/internal/interfaces/http/middleware"
	"gostartkit/pkg/logger"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// New constructs the Gin engine and wires base middlewares and API routes.
func New(userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	applyBaseMiddlewares(r)
	registerHealthRoutes(r)
	configureTrustedProxies(r, cfg)
	attachLoggerAndSecurity(r, cfg)
	applyCORSFromConfig(r, cfg)
	registerAPIV1Routes(r, userHandler, cfg, authMiddleware...)
	return r
}

func applyBaseMiddlewares(r *gin.Engine) {
	r.Use(middleware.JSONRecovery())
	r.Use(middleware.RequestID())
	// Locale detection (Accept-Language → context)
	r.Use(middleware.LocaleMiddleware())
	if err := r.SetTrustedProxies(nil); err != nil {
		logger.L().Warn("set_trusted_proxies_error", "error", err)
	}
}

func configureTrustedProxies(r *gin.Engine, cfg *config.Config) {
	if len(cfg.HTTP.TrustedProxies) == 0 {
		return
	}
	if err := r.SetTrustedProxies(cfg.HTTP.TrustedProxies); err != nil {
		logger.L().Warn("set_trusted_proxies_error", "error", err)
	}
}

func attachLoggerAndSecurity(r *gin.Engine, cfg *config.Config) {
	r.Use(middleware.Logger(logger.L()))
	if cfg.HTTP.SecurityHeaders {
		r.Use(middleware.SecurityHeaders())
	}
}

func applyCORSFromConfig(r *gin.Engine, cfg *config.Config) {
	corsCfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	if len(cfg.HTTP.AllowedOrigins) == 0 || (len(cfg.HTTP.AllowedOrigins) == 1 && cfg.HTTP.AllowedOrigins[0] == "*") {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = cfg.HTTP.AllowedOrigins
	}
	r.Use(cors.New(corsCfg))
}
