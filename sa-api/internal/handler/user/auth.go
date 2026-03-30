package user

import (
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/labstack/echo/v4"
)

// AuthHandler xử lý các API liên quan đến xác thực
type AuthHandler struct {
	userUsecase usecase.UserUsecase
}

// NewAuthHandler tạo instance AuthHandler
func NewAuthHandler(userUsecase usecase.UserUsecase) *AuthHandler {
	return &AuthHandler{userUsecase: userUsecase}
}

// loginRequest request body cho API đăng nhập
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login godoc
// @Summary Đăng nhập hệ thống
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body loginRequest true "Thông tin đăng nhập"
// @Success 200 {object} response.Response{data=usecase.LoginResponse}
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
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

	return response.OKWithMessage(c, "Đăng nhập thành công", result)
}

// Me godoc
// @Summary Lấy thông tin người dùng đang đăng nhập
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=entity.User}
// @Router /auth/me [get]
func (h *AuthHandler) Me(c echo.Context) error {
	userID := getUserIDFromContext(c)
	user, err := h.userUsecase.GetByID(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, user)
}

// ChangePassword godoc
// @Summary Đổi mật khẩu
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.ChangePasswordRequest true "Thông tin đổi mật khẩu"
// @Success 200 {object} response.Response
// @Router /auth/change-password [put]
func (h *AuthHandler) ChangePassword(c echo.Context) error {
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

	userID := getUserIDFromContext(c)
	if err := h.userUsecase.ChangePassword(c.Request().Context(), userID, usecase.ChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}); err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Đổi mật khẩu thành công", nil)
}
