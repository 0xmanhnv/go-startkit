package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

const (
    // RequestIDHeader is the HTTP header used to propagate request correlation ID
    RequestIDHeader   = "X-Request-Id"
    ContextKeyRequestID = "request_id"
)

// RequestID ensures every request has a UUIDv4 correlation ID.
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.GetHeader(RequestIDHeader)
        if id == "" {
            id = uuid.NewString()
        }
        c.Set(ContextKeyRequestID, id)
        c.Writer.Header().Set(RequestIDHeader, id)
        c.Next()
    }
}

