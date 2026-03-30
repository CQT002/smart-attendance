package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// newRequestID tạo request ID ngẫu nhiên dạng hex (không cần thư viện ngoài)
func newRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RequestLogger middleware ghi log mỗi HTTP request với slog
// Bao gồm request_id để trace request end-to-end
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			requestID := newRequestID()

			// Gán request ID vào header để client và downstream service có thể trace
			c.Request().Header.Set("X-Request-ID", requestID)
			c.Response().Header().Set("X-Request-ID", requestID)

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status

			logLevel := slog.LevelInfo
			if status >= 500 {
				logLevel = slog.LevelError
			} else if status >= 400 {
				logLevel = slog.LevelWarn
			}

			slog.Log(c.Request().Context(), logLevel, "http request",
				"request_id", requestID,
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"ip", c.RealIP(),
				"user_agent", c.Request().UserAgent(),
				"user_id", c.Get(ContextKeyUserID),
			)

			return err
		}
	}
}
