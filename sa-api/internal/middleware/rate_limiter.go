package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hdbank/smart-attendance/internal/infrastructure/cache"
	"github.com/labstack/echo/v4"
)

// RateLimitConfig cấu hình rate limiting
type RateLimitConfig struct {
	Requests int           // Số request tối đa
	Window   time.Duration // Trong khoảng thời gian này
	KeyFunc  func(c echo.Context) string // Hàm tạo key để phân biệt client
}

// RateLimiter middleware giới hạn request rate sử dụng Redis sliding window
// Phù hợp với 5000 user đồng thời mà không gây bottleneck
func RateLimiter(redisCache cache.Cache, cfg RateLimitConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := "rate:" + cfg.KeyFunc(c)

			// Increment counter
			count, err := redisCache.Incr(c.Request().Context(), key)
			if err != nil {
				// Nếu Redis lỗi, cho phép request tiếp tục (fail-open)
				slog.Error("rate limiter redis error", "error", err)
				return next(c)
			}

			// Chỉ set TTL lần đầu (khi counter = 1)
			if count == 1 {
				redisCache.Expire(c.Request().Context(), key, cfg.Window)
			}

			// Thêm headers để client biết rate limit
			c.Response().Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.Requests))
			c.Response().Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", maxInt(0, cfg.Requests-int(count))))

			if int(count) > cfg.Requests {
				slog.Warn("rate limit exceeded",
					"key", key,
					"count", count,
					"limit", cfg.Requests,
					"ip", c.RealIP(),
				)
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"success": false,
					"error": map[string]string{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "Quá nhiều yêu cầu, vui lòng thử lại sau",
					},
				})
			}

			return next(c)
		}
	}
}

// CheckInRateLimiter rate limiter đặc biệt cho API check-in/out
// Giới hạn 10 lần/phút per user để chống spam
func CheckInRateLimiter(redisCache cache.Cache) echo.MiddlewareFunc {
	return RateLimiter(redisCache, RateLimitConfig{
		Requests: 10,
		Window:   1 * time.Minute,
		KeyFunc: func(c echo.Context) string {
			userID := GetUserID(c)
			return fmt.Sprintf("checkin:%d", userID)
		},
	})
}

// LoginRateLimiter rate limiter cho API login
// Giới hạn 10 lần/15 phút per IP để chống brute force
func LoginRateLimiter(redisCache cache.Cache) echo.MiddlewareFunc {
	return RateLimiter(redisCache, RateLimitConfig{
		Requests: 10,
		Window:   15 * time.Minute,
		KeyFunc: func(c echo.Context) string {
			return fmt.Sprintf("login:%s", c.RealIP())
		},
	})
}

// GlobalRateLimiter rate limiter toàn cục
// Giới hạn 100 request/phút per IP
func GlobalRateLimiter(redisCache cache.Cache) echo.MiddlewareFunc {
	return RateLimiter(redisCache, RateLimitConfig{
		Requests: 100,
		Window:   1 * time.Minute,
		KeyFunc: func(c echo.Context) string {
			return fmt.Sprintf("global:%s", c.RealIP())
		},
	})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
