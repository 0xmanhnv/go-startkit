package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// simple in-memory rate limiter per IP and path
type limiterKey struct {
	ip   string
	path string
}

type rateStore struct {
	mu   sync.Mutex
	data map[limiterKey]*rate.Limiter
}

func newRateStore() *rateStore { return &rateStore{data: make(map[limiterKey]*rate.Limiter)} }

func (s *rateStore) get(ip, path string, rps rate.Limit, burst int) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := limiterKey{ip: ip, path: path}
	lim, ok := s.data[key]
	if !ok {
		lim = rate.NewLimiter(rps, burst)
		s.data[key] = lim
	}
	return lim
}

// RateLimit limits requests per IP for a given path group (in-memory, single instance).
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	store := newRateStore()
	return func(c *gin.Context) {
		ip := c.ClientIP()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		lim := store.get(ip, path, rate.Limit(rps), burst)
		if !lim.AllowN(time.Now(), 1) {
			// Estimate reset using a token bucket approximation
			// When empty, time until next token ~= 1/rps seconds
			var retryAfterSec int
			if rps > 0 {
				retryAfterSec = int(1.0 / rps)
				if retryAfterSec < 1 {
					retryAfterSec = 1
				}
			} else {
				retryAfterSec = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfterSec))
			c.Header("X-RateLimit-Limit", strconv.FormatFloat(rps, 'f', -1, 64))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(retryAfterSec)*time.Second).Unix(), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too_many_requests"})
			return
		}
		c.Next()
	}
}

// RateLimitForPath applies rate limit only when the request matches the target path (in-memory, single instance).
func RateLimitForPath(targetPath string, rps float64, burst int) gin.HandlerFunc {
	store := newRateStore()
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
		lim := store.get(ip, targetPath, rate.Limit(rps), burst)
		if !lim.AllowN(time.Now(), 1) {
			var retryAfterSec int
			if rps > 0 {
				retryAfterSec = int(1.0 / rps)
				if retryAfterSec < 1 {
					retryAfterSec = 1
				}
			} else {
				retryAfterSec = 1
			}
			c.Header("Retry-After", strconv.Itoa(retryAfterSec))
			c.Header("X-RateLimit-Limit", strconv.FormatFloat(rps, 'f', -1, 64))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Duration(retryAfterSec)*time.Second).Unix(), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too_many_requests"})
			return
		}
		c.Next()
	}
}
