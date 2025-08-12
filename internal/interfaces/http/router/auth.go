package router

import (
	"gostartkit/internal/application/dto"
	"gostartkit/internal/config"
	"gostartkit/internal/infras/ratelimit"
	"gostartkit/internal/interfaces/http/handler"
	"gostartkit/internal/interfaces/http/middleware"
	"gostartkit/pkg/logger"

	"github.com/gin-gonic/gin"
)

func registerAPIV1Routes(r *gin.Engine, userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) {
	v1 := r.Group("/v1")
	registerAuthRoutes(v1, userHandler, cfg, authMiddleware...)
	registerAdminRoutes(v1, authMiddleware...)
}

func registerAuthRoutes(v1 *gin.RouterGroup, userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) {
	auth := v1.Group("/auth")
	auth.POST("/register", middleware.ValidateJSON[dto.CreateUserRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.Register)

	if cfg.HTTP.LoginRateLimitRPS > 0 && cfg.HTTP.LoginRateLimitBurst > 0 {
		if cfg.Env == "prod" {
			logger.L().Warn("login_rate_limit_in_memory", "note", "in-memory limiter is per-instance; consider Redis for multi-instance", "env", cfg.Env)
		}
		login := auth.Group("")
		login.Use(middleware.RateLimitForPath("/v1/auth/login", cfg.HTTP.LoginRateLimitRPS, cfg.HTTP.LoginRateLimitBurst))
		if cfg.RedisAddr != "" {
			rl := ratelimit.NewRedisLimiter(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB).WithFailClosed(cfg.HTTP.LoginRateLimitFailClosed)
			extract := func(c *gin.Context) string {
				v, exists := c.Get("req")
				if !exists {
					return ""
				}
				if req, ok := v.(dto.LoginRequest); ok {
					return req.Email
				}
				return ""
			}
			login.Use(rl.LimitEmail(cfg.HTTP.LoginRateLimitRPS, cfg.HTTP.LoginRateLimitBurst, extract))
		}
		login.POST("/login", middleware.ValidateJSON[dto.LoginRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.Login)
	} else {
		auth.POST("/login", middleware.ValidateJSON[dto.LoginRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.Login)
	}

	if cfg.Security.RefreshEnabled {
		auth.POST("/refresh", userHandler.Refresh)
		auth.POST("/logout", userHandler.Logout)
	}

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
