package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders sets common security-related HTTP headers.
// Enable only where appropriate (some headers may be adjusted behind proxies/CDN).
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "0")
		// Basic CSP (adjust per frontend needs): block mixed content and restrict default src
		if c.Writer.Header().Get("Content-Security-Policy") == "" {
			c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'self'")
		}
		// HSTS: enable only for HTTPS requests (direct TLS or behind proxy)
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}
		c.Next()
	}
}
