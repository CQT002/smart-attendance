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

// CorrectionHandler xử lý API chấm công bù dành cho Admin/Manager
type CorrectionHandler struct {
	correctionUsecase usecase.CorrectionUsecase
}

// NewCorrectionHandler tạo instance CorrectionHandler
func NewCorrectionHandler(correctionUsecase usecase.CorrectionUsecase) *CorrectionHandler {
	return &CorrectionHandler{correctionUsecase: correctionUsecase}
}

// GetList godoc
// @Summary Lấy danh sách yêu cầu chấm công bù (Admin/Manager)
// @Description Manager chỉ xem được yêu cầu của chi nhánh mình
// @Tags Admin - Correction
// @Security BearerAuth
// @Produce json
// @Param status query string false "Lọc theo trạng thái (pending/approved/rejected)"
// @Param page query int false "Trang (mặc định 1)"
// @Param limit query int false "Số bản ghi/trang (mặc định 20)"
// @Success 200 {object} response.Response{data=[]entity.AttendanceCorrection}
// @Router /admin/corrections [get]
func (h *CorrectionHandler) GetList(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	filter := repository.CorrectionFilter{
		Page:  pagination.Page,
		Limit: pagination.Limit,
	}

	if v := c.QueryParam("status"); v != "" {
		filter.Status = entity.CorrectionStatus(v)
	}

	// Manager chỉ xem được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		branchID := middleware.GetBranchID(c)
		filter.BranchID = branchID
	}

	corrections, total, err := h.correctionUsecase.GetList(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, corrections, total, pagination.Page, pagination.Limit)
}

// GetByID godoc
// @Summary Lấy chi tiết yêu cầu chấm công bù
// @Tags Admin - Correction
// @Security BearerAuth
// @Produce json
// @Param id path int true "Correction ID"
// @Success 200 {object} response.Response{data=entity.AttendanceCorrection}
// @Router /admin/corrections/{id} [get]
func (h *CorrectionHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	correction, err := h.correctionUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	// Manager chỉ xem được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		branchID := middleware.GetBranchID(c)
		if branchID != nil && correction.BranchID != *branchID {
			return response.Error(c, apperrors.ErrForbidden)
		}
	}

	return response.OK(c, correction)
}

// Process godoc
// @Summary Duyệt hoặc từ chối yêu cầu chấm công bù
// @Description Manager duyệt: cập nhật AttendanceLog gốc thành present (VALIDATED) trong transaction
// @Tags Admin - Correction
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Correction ID"
// @Param body body usecase.ProcessCorrectionRequest true "Quyết định duyệt/từ chối"
// @Success 200 {object} response.Response{data=entity.AttendanceCorrection}
// @Router /admin/corrections/{id}/process [put]
func (h *CorrectionHandler) Process(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var req usecase.ProcessCorrectionRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	req.CorrectionID = uint(id)
	req.ProcessedByID = middleware.GetUserID(c)

	// Manager chỉ duyệt được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		correction, err := h.correctionUsecase.GetByID(c.Request().Context(), uint(id))
		if err != nil {
			return response.Error(c, err)
		}
		branchID := middleware.GetBranchID(c)
		if branchID != nil && correction.BranchID != *branchID {
			return response.Error(c, apperrors.ErrForbidden)
		}
	}

	result, err := h.correctionUsecase.Process(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Xử lý yêu cầu chấm công bù thành công", result)
}
