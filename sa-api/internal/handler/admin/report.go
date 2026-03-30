package admin

import (
	"strconv"
	"time"

	domainrepo "github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/internal/middleware"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"github.com/labstack/echo/v4"
)

// ReportHandler xử lý các API báo cáo và dashboard
type ReportHandler struct {
	reportUsecase usecase.ReportUsecase
}

// NewReportHandler tạo instance ReportHandler
func NewReportHandler(reportUsecase usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{reportUsecase: reportUsecase}
}

// GetTodayStats godoc
// @Summary Thống kê chấm công hôm nay theo chi nhánh (Dashboard)
// @Description Admin tổng: xem tất cả chi nhánh hoặc lọc theo branch_id. Chi nhánh trưởng: chỉ thấy chi nhánh mình.
// @Tags Admin - Reports
// @Security BearerAuth
// @Produce json
// @Param branch_id query int    false "Lọc theo chi nhánh (chỉ admin tổng mới dùng được)"
// @Param page      query int    false "Trang (default: 1)"
// @Param limit     query int    false "Số bản ghi/trang (default: 20, max: 100)"
// @Success 200 {object} response.Response{data=[]repository.BranchTodayStats}
// @Router /admin/reports/today [get]
func (h *ReportHandler) GetTodayStats(c echo.Context) error {
	pagination := utils.ParsePagination(c)

	filter := usecase.TodayStatsFilter{
		Page:  pagination.Page,
		Limit: pagination.Limit,
	}

	if middleware.IsAdmin(c) {
		// Admin tổng: branch_id là optional filter
		if v := c.QueryParam("branch_id"); v != "" {
			id, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return response.Error(c, apperrors.NewValidationError(map[string]string{
					"branch_id": "branch_id không hợp lệ",
				}))
			}
			uid := uint(id)
			filter.BranchID = &uid
		}
	} else {
		// Manager: bắt buộc dùng chi nhánh của mình, bỏ qua query param
		branchID := middleware.GetBranchID(c)
		if branchID == nil {
			return response.Error(c, apperrors.ErrForbidden)
		}
		filter.BranchID = branchID
	}

	items, total, err := h.reportUsecase.GetTodayBranchStats(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, items, total, pagination.Page, pagination.Limit)
}

// GetTodayEmployees godoc
// @Summary Danh sách chi tiết nhân viên theo trạng thái chấm công hôm nay
// @Description Admin tổng: xem tất cả hoặc lọc theo branch_id. Chi nhánh trưởng: chỉ thấy chi nhánh mình.
// @Description Bao gồm cả nhân viên vắng mặt (chưa chấm công).
// @Tags Admin - Reports
// @Security BearerAuth
// @Produce json
// @Param branch_id query int    false "Lọc theo chi nhánh (admin tổng)"
// @Param status    query string false "Lọc trạng thái: present | late | early_leave | half_day | absent | suspicious (bỏ trống = tất cả)"
// @Param page      query int    false "Trang (default: 1)"
// @Param limit     query int    false "Số bản ghi/trang (default: 20, max: 100)"
// @Success 200 {object} response.Response{data=[]repository.EmployeeTodayDetail}
// @Router /admin/reports/today/employees [get]
func (h *ReportHandler) GetTodayEmployees(c echo.Context) error {
	pagination := utils.ParsePagination(c)

	filter := domainrepo.TodayEmployeeFilter{
		Page:  pagination.Page,
		Limit: pagination.Limit,
	}

	// ── RBAC: branch filter ──
	if middleware.IsAdmin(c) {
		if v := c.QueryParam("branch_id"); v != "" {
			id, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return response.Error(c, apperrors.NewValidationError(map[string]string{
					"branch_id": "branch_id không hợp lệ",
				}))
			}
			uid := uint(id)
			filter.BranchID = &uid
		}
	} else {
		// Manager: bắt buộc dùng chi nhánh của mình
		branchID := middleware.GetBranchID(c)
		if branchID == nil {
			return response.Error(c, apperrors.ErrForbidden)
		}
		filter.BranchID = branchID
	}

	// ── Status filter (whitelist) ──
	if s := c.QueryParam("status"); s != "" {
		status := domainrepo.TodayEmployeeStatus(s)
		switch status {
		case domainrepo.TodayStatusPresent,
			domainrepo.TodayStatusLate,
			domainrepo.TodayStatusEarlyLeave,
			domainrepo.TodayStatusHalfDay,
			domainrepo.TodayStatusAbsent,
			domainrepo.TodayStatusSuspicious:
			filter.Status = status
		default:
			return response.Error(c, apperrors.NewValidationError(map[string]string{
				"status": "status không hợp lệ, chỉ chấp nhận: present, late, early_leave, half_day, absent, suspicious",
			}))
		}
	}

	items, total, err := h.reportUsecase.GetTodayEmployeeDetails(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, items, total, pagination.Page, pagination.Limit)
}

// GetDashboard godoc
// @Summary Lấy thống kê dashboard
// @Description Admin: thống kê toàn hệ thống. Manager: thống kê chi nhánh mình
// @Tags Admin - Reports
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=usecase.DashboardStats}
// @Router /admin/reports/dashboard [get]
func (h *ReportHandler) GetDashboard(c echo.Context) error {
	ctx := c.Request().Context()

	if middleware.IsAdmin(c) {
		stats, err := h.reportUsecase.GetDashboardStats(ctx)
		if err != nil {
			return response.Error(c, err)
		}
		return response.OK(c, stats)
	}

	// Manager: dashboard chi nhánh của mình
	branchID := middleware.GetBranchID(c)
	if branchID == nil {
		return response.Error(c, apperrors.ErrForbidden)
	}

	stats, err := h.reportUsecase.GetBranchDashboard(ctx, *branchID)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, stats)
}

// GetAttendanceReport godoc
// @Summary Báo cáo chấm công chi tiết theo nhân viên
// @Tags Admin - Reports
// @Security BearerAuth
// @Produce json
// @Param period query string false "Chu kỳ: daily|weekly|monthly|custom"
// @Param date_from query string false "Từ ngày (YYYY-MM-DD), dùng khi period=custom"
// @Param date_to query string false "Đến ngày (YYYY-MM-DD), dùng khi period=custom"
// @Param branch_id query int false "Lọc theo chi nhánh"
// @Param department query string false "Lọc theo phòng ban"
// @Param page query int false "Trang"
// @Param limit query int false "Số bản ghi/trang"
// @Success 200 {object} response.Response{data=[]repository.UserAttendanceSummary}
// @Router /admin/reports/attendance [get]
func (h *ReportHandler) GetAttendanceReport(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	filter := buildReportFilter(c, pagination)

	// Manager chỉ xem báo cáo chi nhánh mình
	if !middleware.IsAdmin(c) {
		branchID := middleware.GetBranchID(c)
		filter.BranchID = branchID
	}

	results, total, err := h.reportUsecase.GetAttendanceReport(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, results, total, pagination.Page, pagination.Limit)
}

// GetBranchReport godoc
// @Summary Báo cáo tổng hợp theo chi nhánh (Admin)
// @Tags Admin - Reports
// @Security BearerAuth
// @Produce json
// @Param period query string false "Chu kỳ: daily|weekly|monthly|custom"
// @Param date_from query string false "Từ ngày (YYYY-MM-DD)"
// @Param date_to query string false "Đến ngày (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=[]usecase.BranchAttendanceReport}
// @Router /admin/reports/branches [get]
func (h *ReportHandler) GetBranchReport(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	filter := buildReportFilter(c, pagination)

	results, err := h.reportUsecase.GetBranchReport(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, results)
}

// GetUserReport godoc
// @Summary Báo cáo chấm công chi tiết của một nhân viên
// @Tags Admin - Reports
// @Security BearerAuth
// @Produce json
// @Param user_id path int true "User ID"
// @Param date_from query string true "Từ ngày (YYYY-MM-DD)"
// @Param date_to query string true "Đến ngày (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=repository.UserAttendanceSummary}
// @Router /admin/reports/users/{user_id} [get]
func (h *ReportHandler) GetUserReport(c echo.Context) error {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	from, err := time.Parse("2006-01-02", c.QueryParam("date_from"))
	if err != nil {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"date_from": "Định dạng ngày không hợp lệ (YYYY-MM-DD)",
		}))
	}
	to, err := time.Parse("2006-01-02", c.QueryParam("date_to"))
	if err != nil {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"date_to": "Định dạng ngày không hợp lệ (YYYY-MM-DD)",
		}))
	}

	result, err := h.reportUsecase.GetUserReport(c.Request().Context(), uint(userID), from, to)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, result)
}

// buildReportFilter đọc các tham số filter từ query string
func buildReportFilter(c echo.Context, pagination utils.PaginationParams) usecase.ReportFilter {
	filter := usecase.ReportFilter{
		Period:     usecase.ReportPeriod(c.QueryParam("period")),
		Department: c.QueryParam("department"),
		Page:       pagination.Page,
		Limit:      pagination.Limit,
	}

	if filter.Period == "" {
		filter.Period = usecase.PeriodMonthly
	}

	if v := c.QueryParam("branch_id"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		uid := uint(id)
		filter.BranchID = &uid
	}
	if v := c.QueryParam("date_from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err == nil {
			filter.DateFrom = t
		}
	}
	if v := c.QueryParam("date_to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err == nil {
			filter.DateTo = t
		}
	}

	return filter
}
