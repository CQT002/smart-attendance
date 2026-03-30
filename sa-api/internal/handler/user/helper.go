package user

import (
	"github.com/hdbank/smart-attendance/internal/middleware"
	"github.com/labstack/echo/v4"
)

// getUserIDFromContext lấy user ID từ JWT claims trong context
func getUserIDFromContext(c echo.Context) uint {
	return middleware.GetUserID(c)
}
