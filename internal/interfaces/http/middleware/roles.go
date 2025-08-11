package middleware

import (
	"appsechub/pkg/logger"
	"appsechub/pkg/rbac"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRoles ensures the authenticated user has one of the allowed roles.
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextKeyUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		role, _ := roleVal.(string)
		if _, ok := allowed[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// RequirePermissions checks if the user's role grants any of the required permissions (RBAC policy).
func RequirePermissions(perms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextKeyUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		role, _ := roleVal.(string)
		if !rbac.RoleExists(role) {
			// Warn about unknown role in policy to help misconfig detection
			logger.L().Warn("unknown_role", "role", role, "request_id", c.GetString(ContextKeyRequestID))
		}
		for _, p := range perms {
			if rbac.HasPermission(role, p) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}
