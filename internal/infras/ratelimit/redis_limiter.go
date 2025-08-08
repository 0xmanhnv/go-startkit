package ratelimit

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisLimiter provides a distributed, per-IP-per-path rate limiter using Redis.
// It uses a simple token bucket emulation via INCR with TTL.
type RedisLimiter struct {
	client *redis.Client
}

func NewRedisLimiter(addr, password string, db int) *RedisLimiter {
	return &RedisLimiter{client: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})}
}

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
			// On Redis error, fail-open
			c.Next()
			return
		}
		if count == 1 {
			// First hit in this second, set TTL to 1s
			_ = r.client.Expire(ctx, key, time.Second).Err()
		}
		if int(count) > max {
			c.AbortWithStatusJSON(429, gin.H{"error": "too_many_requests"})
			return
		}
		c.Next()
	}
}
