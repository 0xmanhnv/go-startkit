package http

import (
	"net/http"
	"time"

	"appsechub/internal/application/dto"
	"appsechub/internal/config"
	"appsechub/internal/interfaces/http/handler"
	"appsechub/internal/interfaces/http/middleware"
	"appsechub/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRouter constructs a Gin engine with config (prod-ready). Always require cfg.
func NewRouter(userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) *gin.Engine {
	// Build router from scratch to avoid double route registration
	r := gin.New()
	applyBaseMiddlewares(r)
	registerHealthRoutes(r)
	configureTrustedProxies(r, cfg)
	attachLoggerAndSecurity(r, cfg)
	applyCORSFromConfig(r, cfg)
	registerAPIV1Routes(r, userHandler, cfg, authMiddleware...)
	return r
}

// Helpers
func applyBaseMiddlewares(r *gin.Engine) {
	r.Use(middleware.JSONRecovery())
	r.Use(middleware.RequestID())
	if err := r.SetTrustedProxies(nil); err != nil {
		logger.L().Warn("set_trusted_proxies_error", "error", err)
	}
}

func registerHealthRoutes(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "path": c.Request.URL.Path})
	})
}

func registerAPIV1Routes(r *gin.Engine, userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) {
	v1 := r.Group("/v1")
	registerAuthRoutes(v1, userHandler, cfg, authMiddleware...)
	registerAdminRoutes(v1, authMiddleware...)
}

func registerAuthRoutes(v1 *gin.RouterGroup, userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) {
	auth := v1.Group("/auth")
	// Public routes
	auth.POST(
		"/register",
		middleware.ValidateJSON[dto.CreateUserRequest]("req", cfg.HTTP.MaxBodyBytes),
		userHandler.Register,
	)
	if cfg != nil && cfg.HTTP.LoginRateLimitRPS > 0 && cfg.HTTP.LoginRateLimitBurst > 0 {
		if cfg.Env == "prod" {
			logger.L().Warn("login_rate_limit_in_memory", "note", "in-memory limiter is per-instance; consider Redis for multi-instance", "env", cfg.Env)
		}
		auth.POST("/login",
			middleware.RateLimitForPath("/v1/auth/login", cfg.HTTP.LoginRateLimitRPS, cfg.HTTP.LoginRateLimitBurst),
			middleware.ValidateJSON[dto.LoginRequest]("req", cfg.HTTP.MaxBodyBytes),
			userHandler.Login,
		)
	} else {
		auth.POST("/login", middleware.ValidateJSON[dto.LoginRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.Login)
	}
	// Token endpoints (public by design; protected by token semantics)
	if cfg != nil && cfg.Security.RefreshEnabled {
		auth.POST("/refresh", userHandler.Refresh)
		auth.POST("/logout", userHandler.Logout)
	}
	// Protected routes
	if len(authMiddleware) > 0 {
		protected := auth.Group("")
		protected.Use(authMiddleware...)
		protected.GET("/me", userHandler.GetMe)
		protected.POST("/change-password", middleware.ValidateJSON[dto.ChangePasswordRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.ChangePassword)
		return
	}
	auth.GET("/me", userHandler.GetMe)
	auth.POST("/change-password", middleware.ValidateJSON[dto.ChangePasswordRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.ChangePassword)
}

func registerAdminRoutes(v1 *gin.RouterGroup, authMiddleware ...gin.HandlerFunc) {
	admin := v1.Group("/admin")
	if len(authMiddleware) > 0 {
		admin.Use(authMiddleware...)
	}
	admin.Use(middleware.RequirePermissions("admin:read"))
	admin.GET("/stats", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
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
