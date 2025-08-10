package middleware

import (
    "time"
    "log/slog"
    "github.com/gin-gonic/gin"
)

// Logger produces structured logs for each HTTP request with latency and status.
func Logger(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        latency := time.Since(start)

        requestID, _ := c.Get(ContextKeyRequestID)
        routePath := c.FullPath()
        if routePath == "" {
            routePath = c.Request.URL.Path
        }
        logger.Info("http_request",
            "method", c.Request.Method,
            "path", routePath,
            "status", c.Writer.Status(),
            "latency_ms", latency.Milliseconds(),
            "request_id", requestID,
            "client_ip", c.ClientIP(),
            "user_agent", c.Request.UserAgent(),
            "referer", c.Request.Referer(),
            "x_forwarded_for", c.GetHeader("X-Forwarded-For"),
            "x_real_ip", c.GetHeader("X-Real-Ip"),
        )
    }
}

