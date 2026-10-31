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
		mu          sync.Mutex
		buckets     = make(map[string]*ipRateBucket)
		lastCleanup = time.Now()
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		// Lazy cleanup of expired buckets without background goroutine leaks
		if now.Sub(lastCleanup) > window*2 {
			for k, b := range buckets {
				if now.After(b.resetTime) {
					delete(buckets, k)
				}
			}
			lastCleanup = now
		}

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
