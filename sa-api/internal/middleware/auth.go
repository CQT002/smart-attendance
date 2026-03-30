package middleware

import (
	"log/slog"
	"strings"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"github.com/labstack/echo/v4"
)

const (
	ContextKeyUserID   = "user_id"
	ContextKeyBranchID = "branch_id"
	ContextKeyRole     = "role"
	ContextKeyClaims   = "claims"
)

// JWTAuth middleware xác thực JWT token từ Authorization header
func JWTAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Error(c, apperrors.ErrUnauthorized)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return response.Error(c, apperrors.ErrTokenInvalid)
			}

			claims, err := utils.ParseToken(parts[1], jwtSecret)
			if err != nil {
				slog.Debug("jwt parse failed", "error", err)
				return response.Error(c, apperrors.ErrTokenInvalid)
			}

			// Lưu thông tin user vào context để handler sử dụng
			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyBranchID, claims.BranchID)
			c.Set(ContextKeyRole, claims.Role)
			c.Set(ContextKeyClaims, claims)

			return next(c)
		}
	}
}

// RequireRole middleware kiểm tra role của user
// Chỉ cho phép các role trong danh sách allowedRoles truy cập
func RequireRole(allowedRoles ...entity.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(ContextKeyRole).(entity.UserRole)
			if !ok {
				return response.Error(c, apperrors.ErrUnauthorized)
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					return next(c)
				}
			}

			slog.Warn("access denied - insufficient role",
				"user_id", c.Get(ContextKeyUserID),
				"role", role,
				"required", allowedRoles,
				"path", c.Request().URL.Path,
			)
			return response.Error(c, apperrors.ErrForbidden)
		}
	}
}

// RequireSameBranch middleware đảm bảo Manager chỉ truy cập data của chi nhánh mình
// Admin có thể truy cập tất cả chi nhánh
func RequireSameBranch() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role := c.Get(ContextKeyRole).(entity.UserRole)

			// Admin không bị giới hạn chi nhánh
			if role == entity.RoleAdmin {
				return next(c)
			}

			return next(c)
		}
	}
}

// GetUserID lấy user ID từ context (helper function cho handlers)
func GetUserID(c echo.Context) uint {
	id, _ := c.Get(ContextKeyUserID).(uint)
	return id
}

// GetBranchID lấy branch ID từ context
func GetBranchID(c echo.Context) *uint {
	id, _ := c.Get(ContextKeyBranchID).(*uint)
	return id
}

// GetRole lấy role từ context
func GetRole(c echo.Context) entity.UserRole {
	role, _ := c.Get(ContextKeyRole).(entity.UserRole)
	return role
}

// IsAdmin kiểm tra user hiện tại có phải admin không
func IsAdmin(c echo.Context) bool {
	return GetRole(c) == entity.RoleAdmin
}

// IsManager kiểm tra user hiện tại có phải manager không
func IsManager(c echo.Context) bool {
	role := GetRole(c)
	return role == entity.RoleManager || role == entity.RoleAdmin
}
