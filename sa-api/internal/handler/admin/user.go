package admin

import (
	"strconv"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/internal/middleware"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"github.com/labstack/echo/v4"
)

// UserHandler xử lý các API quản lý người dùng
type UserHandler struct {
	userUsecase usecase.UserUsecase
}

// NewUserHandler tạo instance UserHandler
func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// Create godoc
// @Summary Tạo mới nhân viên (Admin/Manager)
// @Tags Admin - Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.CreateUserRequest true "Thông tin nhân viên"
// @Success 201 {object} response.Response{data=entity.User}
// @Router /admin/users [post]
func (h *UserHandler) Create(c echo.Context) error {
	var req usecase.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if err := validateCreateUser(req); err != nil {
		return response.Error(c, err)
	}

	// Manager chỉ được tạo nhân viên trong chi nhánh của mình
	if middleware.GetRole(c) == entity.RoleManager {
		branchID := middleware.GetBranchID(c)
		req.BranchID = branchID
		req.Role = entity.RoleEmployee // Manager không thể tạo Admin
	}

	user, err := h.userUsecase.Create(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, user)
}

// GetList godoc
// @Summary Lấy danh sách nhân viên
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Trang"
// @Param limit query int false "Số bản ghi/trang"
// @Param branch_id query int false "Lọc theo chi nhánh"
// @Param department query string false "Lọc theo phòng ban"
// @Param role query string false "Lọc theo vai trò (admin/manager/employee)"
// @Param search query string false "Tìm kiếm theo tên, email, mã NV"
// @Success 200 {object} response.Response{data=[]entity.User}
// @Router /admin/users [get]
func (h *UserHandler) GetList(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	filter := repository.UserFilter{
		Department: c.QueryParam("department"),
		Search:     c.QueryParam("search"),
		Page:       pagination.Page,
		Limit:      pagination.Limit,
	}

	if v := c.QueryParam("role"); v != "" {
		filter.Role = entity.UserRole(v)
	}
	if v := c.QueryParam("branch_id"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		uid := uint(id)
		filter.BranchID = &uid
	}

	// Manager chỉ xem nhân viên chi nhánh của mình
	if middleware.GetRole(c) == entity.RoleManager {
		branchID := middleware.GetBranchID(c)
		filter.BranchID = branchID
	}

	users, total, err := h.userUsecase.GetList(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, users, total, pagination.Page, pagination.Limit)
}

// GetByID godoc
// @Summary Lấy thông tin chi tiết nhân viên
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.Response{data=entity.User}
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	user, err := h.userUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, user)
}

// Update godoc
// @Summary Cập nhật thông tin nhân viên
// @Tags Admin - Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body usecase.UpdateUserRequest true "Thông tin cập nhật"
// @Success 200 {object} response.Response{data=entity.User}
// @Router /admin/users/{id} [put]
func (h *UserHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var req usecase.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	user, err := h.userUsecase.Update(c.Request().Context(), uint(id), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Cập nhật thành công", user)
}

// Delete godoc
// @Summary Vô hiệu hóa nhân viên
// @Tags Admin - Users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Router /admin/users/{id} [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if err := h.userUsecase.Delete(c.Request().Context(), uint(id)); err != nil {
		return response.Error(c, err)
	}

	return response.NoContent(c)
}

// ResetPassword godoc
// @Summary Đặt lại mật khẩu nhân viên (Admin/Manager)
// @Tags Admin - Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.Response
// @Router /admin/users/{id}/reset-password [post]
func (h *UserHandler) ResetPassword(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var body struct {
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind(&body); err != nil || body.NewPassword == "" {
		return response.Error(c, apperrors.ErrValidation)
	}
	if len(body.NewPassword) < 6 {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"new_password": "Mật khẩu phải có ít nhất 6 ký tự",
		}))
	}

	if err := h.userUsecase.ResetPassword(c.Request().Context(), uint(id), body.NewPassword); err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Đặt lại mật khẩu thành công", nil)
}

func validateCreateUser(req usecase.CreateUserRequest) error {
	fields := map[string]string{}
	if req.Name == "" {
		fields["name"] = "Tên không được để trống"
	}
	if req.Email == "" {
		fields["email"] = "Email không được để trống"
	}
	if req.Password == "" || len(req.Password) < 6 {
		fields["password"] = "Mật khẩu phải có ít nhất 6 ký tự"
	}
	if req.EmployeeCode == "" {
		fields["employee_code"] = "Mã nhân viên không được để trống"
	}
	if len(fields) > 0 {
		return apperrors.NewValidationError(fields)
	}
	return nil
}
