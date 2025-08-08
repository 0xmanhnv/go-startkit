package http

import (
    "log"
    "net/http"
    "time"

    "appsechub/internal/interfaces/http/handler"
    "appsechub/internal/interfaces/http/middleware"
    "appsechub/internal/config"
    applog "appsechub/pkg/logger"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func NewRouter(userHandler *handler.UserHandler, authMiddleware ...gin.HandlerFunc) *gin.Engine {
    r := gin.New()
    applyBaseMiddlewares(r)
    registerHealthRoutes(r)
    registerAPIV1Routes(r, userHandler, authMiddleware...)
    return r
}

// NewRouterWithConfig allows configuring CORS and trusted proxies from config (prod-ready)
func NewRouterWithConfig(userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) *gin.Engine {
    r := NewRouter(userHandler, authMiddleware...)
    configureTrustedProxies(r, cfg)
    attachLoggerAndSecurity(r, cfg)
    applyCORSFromConfig(r, cfg)
    registerAuthLogin(r, userHandler, cfg)
    return r
}

// Helpers
func applyBaseMiddlewares(r *gin.Engine) {
    r.Use(gin.Recovery())
    r.Use(middleware.RequestID())
    if err := r.SetTrustedProxies(nil); err != nil {
        log.Printf("failed to set trusted proxies: %v", err)
    }
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-Id"},
        AllowCredentials: false,
        MaxAge:           12 * time.Hour,
    }))
}

func registerHealthRoutes(r *gin.Engine) {
    r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
    r.NoRoute(func(c *gin.Context) {
        c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "path": c.Request.URL.Path})
    })
}

func registerAPIV1Routes(r *gin.Engine, userHandler *handler.UserHandler, authMiddleware ...gin.HandlerFunc) {
    v1 := r.Group("/v1")
    registerAuthRegister(v1, userHandler)
    registerAuthMeAndChangePassword(v1, userHandler, authMiddleware...)
    registerAdminRoutes(v1, authMiddleware...)
}

func registerAuthRegister(v1 *gin.RouterGroup, userHandler *handler.UserHandler) {
    auth := v1.Group("/auth")
    auth.POST("/register", userHandler.Register)
}

func registerAuthMeAndChangePassword(v1 *gin.RouterGroup, userHandler *handler.UserHandler, authMiddleware ...gin.HandlerFunc) {
    auth := v1.Group("/auth")
    if len(authMiddleware) > 0 {
        auth.Use(authMiddleware...)
    }
    auth.GET("/me", userHandler.GetMe)
    auth.POST("/change-password", userHandler.ChangePassword)
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
        log.Printf("failed to set trusted proxies: %v", err)
    }
}

func attachLoggerAndSecurity(r *gin.Engine, cfg *config.Config) {
    r.Use(middleware.Logger(applog.New(applog.Options{Level: cfg.LogLevel, Format: "json", AddSource: cfg.Env != "prod"})))
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

func registerAuthLogin(r *gin.Engine, userHandler *handler.UserHandler, cfg *config.Config) {
    auth := r.Group("/v1/auth")
    if cfg.HTTP.LoginRateLimitRPS > 0 && cfg.HTTP.LoginRateLimitBurst > 0 {
        auth.POST("/login", middleware.RateLimitForPath("/v1/auth/login", cfg.HTTP.LoginRateLimitRPS, cfg.HTTP.LoginRateLimitBurst), userHandler.Login)
        return
    }
    auth.POST("/login", userHandler.Login)
}
