package ratelimit

import (
	"fmt"
	"strconv"
	"time"

	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisLimiter provides a distributed, per-IP-per-path rate limiter using Redis.
// It uses a simple token bucket emulation via INCR with TTL.
type RedisLimiter struct {
	client *redis.Client
	// If true, when Redis errors occur, the middleware will deny requests (fail-closed) instead of allowing (fail-open).
	FailClosed bool
}

func NewRedisLimiter(addr, password string, db int) *RedisLimiter {
	return &RedisLimiter{client: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})}
}

// WithFailClosed toggles fail-closed behavior and returns the limiter for chaining.
func (r *RedisLimiter) WithFailClosed(enabled bool) *RedisLimiter { r.FailClosed = enabled; return r }

// Middleware limits requests for a specific path using a windowed counter.
// rps defines requests per second; burst defines allowed burst within the same second.
func (r *RedisLimiter) Middleware(targetPath string, rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 || burst <= 0 { // disabled
		return func(c *gin.Context) { c.Next() }
	}
	// Window size = 1 second; allow up to max = ceil(rps) + burst in that window
	max := int(rps)
	if float64(max) < rps {
		max++
	}
	max += burst
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		if path != targetPath {
			c.Next()
			return
		}
		ip := c.ClientIP()
		now := time.Now().Unix()
		key := fmt.Sprintf("rl:%s:%s:%d", path, ip, now)
		ctx := c.Request.Context()
		count, err := r.client.Incr(ctx, key).Result()
		if err != nil {
			if r.FailClosed {
				c.AbortWithStatusJSON(429, gin.H{"error": gin.H{"code": "too_many_requests", "message": "too many requests"}})
				return
			}
			// fail-open
			c.Next()
			return
		}
		if count == 1 {
			// First hit in this second, set TTL to 1s
			_ = r.client.Expire(ctx, key, time.Second).Err()
		}
		if int(count) > max {
			// Compute reset at the end of the current 1s window
			reset := time.Unix(now+1, 0)
			c.Header("Retry-After", "1")
			c.Header("X-RateLimit-Limit", strconv.Itoa(max))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
			c.AbortWithStatusJSON(429, gin.H{"error": gin.H{"code": "too_many_requests", "message": "too many requests"}})
			return
		}
		c.Next()
	}
}

// MiddlewareWithEmail adds per-email rate limit in addition to IP if email present under c.Keys[keyName].
// Use together with a binding middleware that stores the DTO into context.
func (r *RedisLimiter) MiddlewareWithEmail(targetPath string, rps float64, burst int, email string) gin.HandlerFunc {
	base := r.Middleware(targetPath, rps, burst)
	if rps <= 0 || burst <= 0 {
		return base
	}
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		if path != targetPath || email == "" {
			base(c)
			return
		}
		now := time.Now().Unix()
		key := fmt.Sprintf("rl:%s:email:%s:%d", path, email, now)
		ctx := c.Request.Context()
		max := int(rps)
		if float64(max) < rps {
			max++
		}
		max += burst
		count, err := r.client.Incr(ctx, key).Result()
		if err != nil {
			if r.FailClosed {
				c.AbortWithStatusJSON(429, gin.H{"error": gin.H{"code": "too_many_requests", "message": "too many requests"}})
				return
			}
			base(c)
			return
		}
		if count == 1 {
			_ = r.client.Expire(ctx, key, time.Second).Err()
		}
		if int(count) > max {
			reset := time.Unix(now+1, 0)
			c.Header("Retry-After", "1")
			c.Header("X-RateLimit-Limit", strconv.Itoa(max))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
			c.AbortWithStatusJSON(429, gin.H{"error": gin.H{"code": "too_many_requests", "message": "too many requests"}})
			return
		}
		base(c)
	}
}

// LimitEmail applies rate limit using an email extractor function. Intended for route-level use
// after JSON binding middleware has populated context with the request DTO.
func (r *RedisLimiter) LimitEmail(rps float64, burst int, extractEmail func(*gin.Context) string) gin.HandlerFunc {
	if rps <= 0 || burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		email := ""
		if extractEmail != nil {
			email = extractEmail(c)
		}
		email = strings.TrimSpace(strings.ToLower(email))
		if email == "" {
			c.Next()
			return
		}
		sum := sha256.Sum256([]byte(email))
		emailKey := hex.EncodeToString(sum[:])
		now := time.Now().Unix()
		max := int(rps)
		if float64(max) < rps {
			max++
		}
		max += burst
		key := fmt.Sprintf("rl:email:%s:%d", emailKey, now)
		ctx := c.Request.Context()
		count, err := r.client.Incr(ctx, key).Result()
		if err != nil {
			if r.FailClosed {
				c.AbortWithStatusJSON(429, gin.H{"error": gin.H{"code": "too_many_requests", "message": "too many requests"}})
				return
			}
			c.Next()
			return
		}
		if count == 1 {
			_ = r.client.Expire(ctx, key, time.Second).Err()
		}
		if int(count) > max {
			reset := time.Unix(now+1, 0)
			c.Header("Retry-After", "1")
			c.Header("X-RateLimit-Limit", strconv.Itoa(max))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
			c.AbortWithStatusJSON(429, gin.H{"error": gin.H{"code": "too_many_requests", "message": "too many requests"}})
			return
		}
		c.Next()
	}
}
