package router

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipRateBucket struct {
	count     int
	resetTime time.Time
}

// RateLimit returns a Gin middleware that limits requests per client IP
// within the given time window.
func RateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	var (
		mu      sync.Mutex
		buckets = make(map[string]*ipRateBucket)
	)

	// Periodic background cleanup of expired IP buckets
	go func() {
		ticker := time.NewTicker(window * 2)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			mu.Lock()
			for ip, b := range buckets {
				if now.After(b.resetTime) {
					delete(buckets, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		b, exists := buckets[ip]
		if !exists || now.After(b.resetTime) {
			buckets[ip] = &ipRateBucket{
				count:     1,
				resetTime: now.Add(window),
			}
			mu.Unlock()
			c.Next()
			return
		}

		if b.count >= maxRequests {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, ErrorResponse{
				Error: "too many requests, please try again later",
			})
			return
		}

		b.count++
		mu.Unlock()
		c.Next()
	}
}

