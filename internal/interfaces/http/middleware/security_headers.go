package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets common security-related HTTP headers.
// Enable only where appropriate (some headers may be adjusted behind proxies/CDN).
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
        c.Writer.Header().Set("Referrer-Policy", "no-referrer")
        c.Writer.Header().Set("X-Frame-Options", "DENY")
        // HSTS: enable only for HTTPS deployments; adjust max-age as needed
        // If running behind TLS-terminating proxy, you may want to set via proxy
        c.Writer.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
        c.Next()
    }
}

