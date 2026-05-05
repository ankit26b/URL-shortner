package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	redisClient *redis.Client
	limit       int64
	window      time.Duration
}

func NewRateLimiter(redisClient *redis.Client, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		limit:       limit,
		window:      window,
	}
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if r.redisClient == nil || r.limit <= 0 || r.window <= 0 {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:shorten:%s", ip)

		count, err := r.redisClient.Incr(c.Request.Context(), key).Result()
		if err != nil {
			c.Next()
			return
		}

		if count == 1 {
			_ = r.redisClient.Expire(c.Request.Context(), key, r.window).Err()
		}

		if count > r.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
