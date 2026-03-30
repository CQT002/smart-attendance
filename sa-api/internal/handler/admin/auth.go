package admin

import (
	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/internal/middleware"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/labstack/echo/v4"
)

// AdminAuthHandler xử lý xác thực cho Admin Portal
type AdminAuthHandler struct {
	userUsecase usecase.UserUsecase
}

// NewAdminAuthHandler tạo instance AdminAuthHandler
func NewAdminAuthHandler(userUsecase usecase.UserUsecase) *AdminAuthHandler {
	return &AdminAuthHandler{userUsecase: userUsecase}
}

// adminLoginRequest request body đăng nhập admin portal
type adminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login godoc
// @Summary Đăng nhập Admin Portal
// @Description Chỉ cho phép tài khoản có role admin hoặc manager. Employee bị từ chối.
// @Tags Admin - Auth
// @Accept json
// @Produce json
// @Param body body adminLoginRequest true "Thông tin đăng nhập"
// @Success 200 {object} response.Response{data=usecase.LoginResponse}
// @Failure 401 {object} response.Response "Sai email hoặc mật khẩu"
// @Failure 403 {object} response.Response "Tài khoản không có quyền truy cập Admin Portal"
// @Router /admin/auth/login [post]
func (h *AdminAuthHandler) Login(c echo.Context) error {
	var req adminLoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if req.Email == "" || req.Password == "" {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"email":    "Email không được để trống",
			"password": "Mật khẩu không được để trống",
		}))
	}

	result, err := h.userUsecase.Login(c.Request().Context(), usecase.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return response.Error(c, err)
	}

	// Chỉ cho phép admin và manager truy cập Admin Portal
	if result.User.Role == entity.RoleEmployee {
		return response.Error(c, apperrors.New(403, "FORBIDDEN", "Tài khoản không có quyền truy cập Admin Portal"))
	}

	return response.OKWithMessage(c, "Đăng nhập Admin Portal thành công", result)
}

// Me godoc
// @Summary Lấy thông tin admin/manager đang đăng nhập
// @Tags Admin - Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=entity.User}
// @Router /admin/auth/me [get]
func (h *AdminAuthHandler) Me(c echo.Context) error {
	userID := middleware.GetUserID(c)
	user, err := h.userUsecase.GetByID(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, user)
}

// ChangePassword godoc
// @Summary Đổi mật khẩu (Admin Portal)
// @Tags Admin - Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.ChangePasswordRequest true "Thông tin đổi mật khẩu"
// @Success 200 {object} response.Response
// @Router /admin/auth/change-password [put]
func (h *AdminAuthHandler) ChangePassword(c echo.Context) error {
	var req usecase.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return response.Error(c, apperrors.ErrValidation)
	}
	if len(req.NewPassword) < 6 {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"new_password": "Mật khẩu mới phải có ít nhất 6 ký tự",
		}))
	}

	userID := middleware.GetUserID(c)
	if err := h.userUsecase.ChangePassword(c.Request().Context(), userID, req); err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Đổi mật khẩu thành công", nil)
}
