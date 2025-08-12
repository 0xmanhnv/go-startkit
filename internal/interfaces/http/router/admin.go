package router

import (
	"gostartkit/internal/interfaces/http/middleware"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(v1 *gin.RouterGroup, authMiddleware ...gin.HandlerFunc) {
	admin := v1.Group("/admin")
	if len(authMiddleware) > 0 {
		admin.Use(authMiddleware...)
	}
	admin.Use(middleware.RequirePermissions("admin:read"))
	admin.GET("/stats", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
}
