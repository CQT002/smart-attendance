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

// OvertimeHandler xử lý API tăng ca dành cho Admin/Manager
type OvertimeHandler struct {
	overtimeUsecase usecase.OvertimeUsecase
}

// NewOvertimeHandler tạo instance OvertimeHandler
func NewOvertimeHandler(overtimeUsecase usecase.OvertimeUsecase) *OvertimeHandler {
	return &OvertimeHandler{overtimeUsecase: overtimeUsecase}
}

// GetList godoc
// @Summary Lấy danh sách yêu cầu tăng ca (Admin/Manager)
// @Description Manager chỉ xem được yêu cầu của chi nhánh mình
// @Tags Admin - Overtime
// @Security BearerAuth
// @Produce json
// @Param status query string false "Lọc theo trạng thái (pending/approved/rejected)"
// @Param page query int false "Trang (mặc định 1)"
// @Param limit query int false "Số bản ghi/trang (mặc định 20)"
// @Success 200 {object} response.Response{data=[]entity.OvertimeRequest}
// @Router /admin/overtime [get]
func (h *OvertimeHandler) GetList(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	filter := repository.OvertimeFilter{
		Page:  pagination.Page,
		Limit: pagination.Limit,
	}

	if v := c.QueryParam("status"); v != "" {
		filter.Status = entity.OvertimeStatus(v)
	}

	// Manager chỉ xem được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		branchID := middleware.GetBranchID(c)
		filter.BranchID = branchID
	}

	overtimes, total, err := h.overtimeUsecase.GetList(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, overtimes, total, pagination.Page, pagination.Limit)
}

// GetByID godoc
// @Summary Lấy chi tiết yêu cầu tăng ca
// @Tags Admin - Overtime
// @Security BearerAuth
// @Produce json
// @Param id path int true "Overtime ID"
// @Success 200 {object} response.Response{data=entity.OvertimeRequest}
// @Router /admin/overtime/{id} [get]
func (h *OvertimeHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	ot, err := h.overtimeUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	// Manager chỉ xem được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		branchID := middleware.GetBranchID(c)
		if branchID != nil && ot.BranchID != *branchID {
			return response.Error(c, apperrors.ErrForbidden)
		}
	}

	return response.OK(c, ot)
}

// Process godoc
// @Summary Duyệt hoặc từ chối yêu cầu tăng ca
// @Description Manager duyệt: tính calculated_start/end/total_hours trong transaction
// @Tags Admin - Overtime
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Overtime ID"
// @Param body body usecase.ProcessOvertimeRequest true "Quyết định duyệt/từ chối"
// @Success 200 {object} response.Response{data=entity.OvertimeRequest}
// @Router /admin/overtime/{id}/process [put]
func (h *OvertimeHandler) Process(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var req usecase.ProcessOvertimeRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	req.OvertimeID = uint(id)
	req.ProcessedByID = middleware.GetUserID(c)

	// Manager chỉ duyệt được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		ot, err := h.overtimeUsecase.GetByID(c.Request().Context(), uint(id))
		if err != nil {
			return response.Error(c, err)
		}
		branchID := middleware.GetBranchID(c)
		if branchID != nil && ot.BranchID != *branchID {
			return response.Error(c, apperrors.ErrForbidden)
		}
	}

	result, err := h.overtimeUsecase.Process(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Xử lý yêu cầu tăng ca thành công", result)
}

// BatchApprove godoc
// @Summary Duyệt tất cả yêu cầu tăng ca đang chờ
// @Tags Admin - Overtime
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /admin/overtime/batch-approve [post]
func (h *OvertimeHandler) BatchApprove(c echo.Context) error {
	processedByID := middleware.GetUserID(c)

	var branchID *uint
	if middleware.GetRole(c) == entity.RoleManager {
		branchID = middleware.GetBranchID(c)
	}

	count, err := h.overtimeUsecase.BatchApprove(c.Request().Context(), processedByID, branchID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Đã duyệt hàng loạt", map[string]int64{"approved_count": count})
}
