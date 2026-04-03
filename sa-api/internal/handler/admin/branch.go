package admin

import (
	"strconv"

	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/internal/middleware"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"github.com/labstack/echo/v4"
)

// BranchHandler xử lý các API quản lý chi nhánh
type BranchHandler struct {
	branchUsecase usecase.BranchUsecase
}

// NewBranchHandler tạo instance BranchHandler
func NewBranchHandler(branchUsecase usecase.BranchUsecase) *BranchHandler {
	return &BranchHandler{branchUsecase: branchUsecase}
}

// Create godoc
// @Summary Tạo mới chi nhánh (Admin)
// @Tags Admin - Branches
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.CreateBranchRequest true "Thông tin chi nhánh"
// @Success 201 {object} response.Response{data=entity.Branch}
// @Router /admin/branches [post]
func (h *BranchHandler) Create(c echo.Context) error {
	var req usecase.CreateBranchRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if req.Code == "" || req.Name == "" {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"code": "Mã chi nhánh không được để trống",
			"name": "Tên chi nhánh không được để trống",
		}))
	}

	branch, err := h.branchUsecase.Create(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, branch)
}

// GetList godoc
// @Summary Lấy danh sách chi nhánh
// @Tags Admin - Branches
// @Security BearerAuth
// @Produce json
// @Param page query int false "Trang"
// @Param limit query int false "Số bản ghi/trang"
// @Param search query string false "Tìm kiếm theo tên hoặc mã"
// @Success 200 {object} response.Response{data=[]entity.Branch}
// @Router /admin/branches [get]
func (h *BranchHandler) GetList(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	filter := repository.BranchFilter{
		Search: c.QueryParam("search"),
		Page:   pagination.Page,
		Limit:  pagination.Limit,
	}

	// Manager: chỉ xem chi nhánh của mình
	if !middleware.IsAdmin(c) {
		filter.BranchID = middleware.GetBranchID(c)
	}

	branches, total, err := h.branchUsecase.GetList(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, branches, total, pagination.Page, pagination.Limit)
}

// GetActive godoc
// @Summary Lấy danh sách chi nhánh đang hoạt động
// @Tags Admin - Branches
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]entity.Branch}
// @Router /admin/branches/active [get]
func (h *BranchHandler) GetActive(c echo.Context) error {
	branches, err := h.branchUsecase.GetActive(c.Request().Context())
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, branches)
}

// GetByID godoc
// @Summary Lấy thông tin chi tiết chi nhánh
// @Tags Admin - Branches
// @Security BearerAuth
// @Produce json
// @Param id path int true "Branch ID"
// @Success 200 {object} response.Response{data=entity.Branch}
// @Router /admin/branches/{id} [get]
func (h *BranchHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	branch, err := h.branchUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, branch)
}

// Update godoc
// @Summary Cập nhật thông tin chi nhánh (Admin)
// @Tags Admin - Branches
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Branch ID"
// @Param body body usecase.UpdateBranchRequest true "Thông tin cập nhật"
// @Success 200 {object} response.Response{data=entity.Branch}
// @Router /admin/branches/{id} [put]
func (h *BranchHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var req usecase.UpdateBranchRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	branch, err := h.branchUsecase.Update(c.Request().Context(), uint(id), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Cập nhật thành công", branch)
}

// Delete godoc
// @Summary Vô hiệu hóa chi nhánh (Admin)
// @Tags Admin - Branches
// @Security BearerAuth
// @Produce json
// @Param id path int true "Branch ID"
// @Success 204
// @Router /admin/branches/{id} [delete]
func (h *BranchHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if err := h.branchUsecase.Delete(c.Request().Context(), uint(id)); err != nil {
		return response.Error(c, err)
	}

	return response.NoContent(c)
}
