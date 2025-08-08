package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID    = "user_id"
	ContextKeyUserRole  = "user_role"
	ContextKeyJWTClaims = "jwt_claims"
)

// TokenValidator validates a token string and returns subject (userID) and role.
type TokenValidator func(token string) (subject string, role string, err error)

func JWTAuth(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_request", error_description="missing Authorization header"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		tokenStr := extractBearerToken(authHeader)
		if tokenStr == "" {
			c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_request", error_description="expected Bearer token"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		subject, role, err := validator(tokenStr)
		if err != nil {
			c.Header("WWW-Authenticate", `Bearer realm="api", error="invalid_token", error_description="token invalid or expired"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(ContextKeyUserID, subject)
		c.Set(ContextKeyUserRole, role)
		c.Next()
	}
}

func extractBearerToken(authorizationHeader string) string {
	// Case-insensitive prefix match, allow extra spaces
	// Examples: "Bearer <token>", "bearer <token>", "Bearer    <token>"
	parts := strings.Fields(authorizationHeader)
	if len(parts) < 2 {
		return ""
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
