package middleware

import (
	"net/http"
	"strings"

	"appsechub/internal/infras/security"
	"github.com/gin-gonic/gin"
)

func JWTAuth(jwtSvc security.JWTService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := jwtSvc.ValidateToken(tokenStr)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

        // Đưa thông tin claims vào context
        c.Set("user_id", claims.Subject)
        c.Next()
    }
}