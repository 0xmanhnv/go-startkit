package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// JSONRecovery recovers from panics and returns a standardized JSON error envelope.
// It avoids leaking internal details to the client while preserving request context.
func JSONRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Best effort: do not expose internal panic details to client
				// Attach request id if available (already added by RequestID middleware)
				reqID := c.Writer.Header().Get(RequestIDHeader)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "server_error",
						"message": "internal error",
					},
					"meta": gin.H{
						"request_id": reqID,
					},
				})
				// Optionally log stack for operators; logging is handled by outer middleware/logger
				_ = debug.Stack()
			}
		}()
		c.Next()
	}
}
