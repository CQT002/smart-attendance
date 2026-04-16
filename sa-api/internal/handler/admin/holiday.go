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

// HolidayHandler xử lý API quản lý ngày lễ cho Admin/Manager.
//   - Admin: full CRUD
//   - Manager: chỉ GET
type HolidayHandler struct {
	holidayUsecase usecase.HolidayUsecase
}

// NewHolidayHandler tạo instance HolidayHandler
func NewHolidayHandler(holidayUsecase usecase.HolidayUsecase) *HolidayHandler {
	return &HolidayHandler{holidayUsecase: holidayUsecase}
}

// GetList godoc
// @Summary Lấy danh sách ngày lễ theo năm (Admin/Manager)
// @Tags Admin - Holiday
// @Security BearerAuth
// @Produce json
// @Param year query int false "Năm (mặc định năm hiện tại)"
// @Param type query string false "Loại (national|company)"
// @Param page query int false "Trang"
// @Param limit query int false "Số bản ghi/trang (mặc định 100)"
// @Success 200 {object} response.Response{data=[]entity.Holiday}
// @Router /admin/holidays [get]
func (h *HolidayHandler) GetList(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	if pagination.Limit == 0 {
		pagination.Limit = 100
	}

	filter := repository.HolidayFilter{
		Page:  pagination.Page,
		Limit: pagination.Limit,
	}

	if v := c.QueryParam("year"); v != "" {
		y, err := strconv.Atoi(v)
		if err != nil || y < 1900 || y > 3000 {
			return response.Error(c, apperrors.NewValidationError(map[string]string{"year": "Năm không hợp lệ"}))
		}
		filter.Year = &y
	} else {
		y := utils.Now().Year()
		filter.Year = &y
	}

	if v := c.QueryParam("type"); v != "" {
		t := entity.HolidayType(v)
		if t != entity.HolidayTypeNational && t != entity.HolidayTypeCompany {
			return response.Error(c, apperrors.ErrHolidayInvalidType)
		}
		filter.Type = t
	}

	holidays, total, err := h.holidayUsecase.GetList(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, holidays, total, pagination.Page, pagination.Limit)
}

// GetByID godoc
// @Summary Lấy chi tiết ngày lễ
// @Tags Admin - Holiday
// @Security BearerAuth
// @Produce json
// @Param id path int true "Holiday ID"
// @Success 200 {object} response.Response{data=entity.Holiday}
// @Router /admin/holidays/{id} [get]
func (h *HolidayHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	holiday, err := h.holidayUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, holiday)
}

// Create godoc
// @Summary Tạo ngày lễ (Admin only)
// @Tags Admin - Holiday
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.CreateHolidayRequest true "Thông tin ngày lễ"
// @Success 200 {object} response.Response{data=entity.Holiday}
// @Router /admin/holidays [post]
func (h *HolidayHandler) Create(c echo.Context) error {
	var req usecase.CreateHolidayRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}
	req.CreatedByID = middleware.GetUserID(c)

	result, err := h.holidayUsecase.Create(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OKWithMessage(c, "Tạo ngày lễ thành công", result)
}

// Update godoc
// @Summary Cập nhật ngày lễ (Admin only)
// @Tags Admin - Holiday
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Holiday ID"
// @Param body body usecase.UpdateHolidayRequest true "Thông tin cập nhật"
// @Success 200 {object} response.Response{data=entity.Holiday}
// @Router /admin/holidays/{id} [put]
func (h *HolidayHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var req usecase.UpdateHolidayRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}
	req.ID = uint(id)

	result, err := h.holidayUsecase.Update(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OKWithMessage(c, "Cập nhật ngày lễ thành công", result)
}

// Delete godoc
// @Summary Xoá ngày lễ (Admin only)
// @Tags Admin - Holiday
// @Security BearerAuth
// @Produce json
// @Param id path int true "Holiday ID"
// @Success 200 {object} response.Response
// @Router /admin/holidays/{id} [delete]
func (h *HolidayHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if err := h.holidayUsecase.Delete(c.Request().Context(), uint(id)); err != nil {
		return response.Error(c, err)
	}
	return response.OKWithMessage(c, "Xoá ngày lễ thành công", nil)
}
