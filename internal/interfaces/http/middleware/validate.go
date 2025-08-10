package middleware

import (
	"net/http"

	"gostartkit/internal/interfaces/http/validation"

	"github.com/gin-gonic/gin"
)

// ValidateJSON binds and validates request JSON into a typed payload before reaching the handler.
// The validated payload is stored in context with the provided ctxKey.
func ValidateJSON[T any](ctxKey string, maxBodyBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBodyBytes > 0 && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
		}
		var payload T
		if err := c.ShouldBindJSON(&payload); err != nil {
			if validation.IsBodyTooLarge(err) {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "payload_too_large", "message": "request body exceeds limit"}})
				return
			}
			code, msg := validation.MapBindJSONError(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": code, "message": msg}})
			return
		}
		c.Set(ctxKey, payload)
		c.Next()
	}
}
