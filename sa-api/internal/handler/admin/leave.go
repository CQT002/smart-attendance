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

// LeaveHandler xử lý API nghỉ phép dành cho Admin/Manager
type LeaveHandler struct {
	leaveUsecase usecase.LeaveUsecase
}

// NewLeaveHandler tạo instance LeaveHandler
func NewLeaveHandler(leaveUsecase usecase.LeaveUsecase) *LeaveHandler {
	return &LeaveHandler{leaveUsecase: leaveUsecase}
}

// GetList godoc
// @Summary Lấy danh sách yêu cầu nghỉ phép (Admin/Manager)
// @Description Manager chỉ xem được yêu cầu của chi nhánh mình
// @Tags Admin - Leave
// @Security BearerAuth
// @Produce json
// @Param status query string false "Lọc theo trạng thái (pending/approved/rejected)"
// @Param page query int false "Trang (mặc định 1)"
// @Param limit query int false "Số bản ghi/trang (mặc định 20)"
// @Success 200 {object} response.Response{data=[]entity.LeaveRequest}
// @Router /admin/leaves [get]
func (h *LeaveHandler) GetList(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	filter := repository.LeaveFilter{
		Page:  pagination.Page,
		Limit: pagination.Limit,
	}

	if v := c.QueryParam("status"); v != "" {
		filter.Status = entity.LeaveStatus(v)
	}

	// Manager chỉ xem được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		branchID := middleware.GetBranchID(c)
		filter.BranchID = branchID
	}

	leaves, total, err := h.leaveUsecase.GetList(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, leaves, total, pagination.Page, pagination.Limit)
}

// GetByID godoc
// @Summary Lấy chi tiết yêu cầu nghỉ phép
// @Tags Admin - Leave
// @Security BearerAuth
// @Produce json
// @Param id path int true "Leave ID"
// @Success 200 {object} response.Response{data=entity.LeaveRequest}
// @Router /admin/leaves/{id} [get]
func (h *LeaveHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	leave, err := h.leaveUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	// Manager chỉ xem được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		branchID := middleware.GetBranchID(c)
		if branchID != nil && leave.BranchID != *branchID {
			return response.Error(c, apperrors.ErrForbidden)
		}
	}

	return response.OK(c, leave)
}

// Process godoc
// @Summary Duyệt hoặc từ chối yêu cầu nghỉ phép
// @Description Manager duyệt: tạo/cập nhật AttendanceLog với status=leave trong transaction
// @Tags Admin - Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Leave ID"
// @Param body body usecase.ProcessLeaveRequest true "Quyết định duyệt/từ chối"
// @Success 200 {object} response.Response{data=entity.LeaveRequest}
// @Router /admin/leaves/{id}/process [put]
func (h *LeaveHandler) Process(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var req usecase.ProcessLeaveRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	req.LeaveID = uint(id)
	req.ProcessedByID = middleware.GetUserID(c)

	// Manager chỉ duyệt được yêu cầu của chi nhánh mình
	if middleware.GetRole(c) == entity.RoleManager {
		leave, err := h.leaveUsecase.GetByID(c.Request().Context(), uint(id))
		if err != nil {
			return response.Error(c, err)
		}
		branchID := middleware.GetBranchID(c)
		if branchID != nil && leave.BranchID != *branchID {
			return response.Error(c, apperrors.ErrForbidden)
		}
	}

	result, err := h.leaveUsecase.Process(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Xử lý yêu cầu nghỉ phép thành công", result)
}

// BatchApprove godoc
// @Summary Duyệt tất cả yêu cầu nghỉ phép đang chờ
// @Tags Admin - Leave
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /admin/leaves/batch-approve [post]
func (h *LeaveHandler) BatchApprove(c echo.Context) error {
	processedByID := middleware.GetUserID(c)

	var branchID *uint
	if middleware.GetRole(c) == entity.RoleManager {
		branchID = middleware.GetBranchID(c)
	}

	count, err := h.leaveUsecase.BatchApprove(c.Request().Context(), processedByID, branchID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Đã duyệt hàng loạt", map[string]int64{"approved_count": count})
}

// GetPendingApprovals godoc
// @Summary Lấy danh sách tổng hợp chờ duyệt (bù công + nghỉ phép)
// @Tags Admin - Approvals
// @Security BearerAuth
// @Produce json
// @Param page query int false "Trang (mặc định 1)"
// @Param limit query int false "Số bản ghi/trang (mặc định 20)"
// @Success 200 {object} response.Response{data=[]usecase.PendingApprovalItem}
// @Router /admin/approvals/pending [get]
func (h *LeaveHandler) GetPendingApprovals(c echo.Context) error {
	pagination := utils.ParsePagination(c)

	var branchID *uint
	if middleware.GetRole(c) == entity.RoleManager {
		branchID = middleware.GetBranchID(c)
	}

	items, total, err := h.leaveUsecase.GetPendingApprovals(c.Request().Context(), branchID, pagination.Page, pagination.Limit)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, items, total, pagination.Page, pagination.Limit)
}

// GetApprovals godoc
// @Summary Lấy danh sách tổng hợp duyệt chấm công (bù công + nghỉ phép) — hỗ trợ lọc theo status
// @Tags Admin - Approvals
// @Security BearerAuth
// @Produce json
// @Param status query string false "Lọc theo trạng thái (pending/approved/rejected)"
// @Param page query int false "Trang (mặc định 1)"
// @Param limit query int false "Số bản ghi/trang (mặc định 20)"
// @Success 200 {object} response.Response{data=[]usecase.PendingApprovalItem}
// @Router /admin/approvals [get]
func (h *LeaveHandler) GetApprovals(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	status := c.QueryParam("status")

	var branchID *uint
	if middleware.GetRole(c) == entity.RoleManager {
		branchID = middleware.GetBranchID(c)
	}

	items, total, err := h.leaveUsecase.GetApprovals(c.Request().Context(), branchID, status, pagination.Page, pagination.Limit)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, items, total, pagination.Page, pagination.Limit)
}
