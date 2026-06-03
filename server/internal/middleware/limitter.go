package middleware

import (
	"context"
	"fmt"

	"server/internal/config"
	"server/internal/config/constant"
	"server/pkg/cache"
	"server/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter menggunakan Redis incr+expire, setara limitter() di Hono kamu.
// Key: rl:<userID_or_IP>:<path>
func RateLimiter(cfg *config.Config, cache *cache.Client, limit constant.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Identifikasi requester: gunakan userID jika sudah auth, fallback ke IP
		identifier := c.ClientIP()
		if user := GetAuthUser(c); user != nil {
			identifier = user.UserID
		}

		key := fmt.Sprintf("%s:%s:%s", cfg.Redis.RateLimitPrefix, identifier, c.FullPath())

		count, err := cache.Incr(context.Background(), key)
		if err != nil {
			// Jika Redis error, allow request (fail open) agar tidak block semua user
			c.Next()
			return
		}

		// Set TTL hanya saat counter baru dibuat
		if count == 1 {
			cache.Expire(context.Background(), key, limit.WindowSec)
		}

		if int(count) > limit.Limit {
			response.TooManyRequests(c,
				constant.ErrTooManyRequests,
				constant.CodeTooManyRequests,
			)
			c.Abort()
			return
		}

		// Set header informatif ke client
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit.Limit-int(count)))

		c.Next()
	}
}