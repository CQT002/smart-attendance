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

// AttendanceHandler xử lý các API chấm công dành cho Admin/Manager
type AttendanceHandler struct {
	attendanceUsecase usecase.AttendanceUsecase
}

// NewAttendanceHandler tạo instance AttendanceHandler
func NewAttendanceHandler(attendanceUsecase usecase.AttendanceUsecase) *AttendanceHandler {
	return &AttendanceHandler{attendanceUsecase: attendanceUsecase}
}

// GetList godoc
// @Summary Lấy danh sách chấm công (Admin/Manager)
// @Description Hỗ trợ filter theo ngày/tuần/tháng, phòng ban, chi nhánh với phân trang
// @Tags Admin - Attendance
// @Security BearerAuth
// @Produce json
// @Param page query int false "Trang (mặc định 1)"
// @Param limit query int false "Số bản ghi/trang (mặc định 20, tối đa 100)"
// @Param branch_id query int false "Lọc theo chi nhánh"
// @Param user_id query int false "Lọc theo nhân viên"
// @Param department query string false "Lọc theo phòng ban"
// @Param status query string false "Lọc theo trạng thái (present/late/absent/early_leave)"
// @Param date_from query string false "Từ ngày (YYYY-MM-DD)"
// @Param date_to query string false "Đến ngày (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=[]entity.AttendanceLog}
// @Router /admin/attendance [get]
func (h *AttendanceHandler) GetList(c echo.Context) error {
	pagination := utils.ParsePagination(c)
	filter := repository.AttendanceFilter{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Department: c.QueryParam("department"),
	}

	if v := c.QueryParam("branch_id"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		uid := uint(id)
		filter.BranchID = &uid
	}
	if v := c.QueryParam("user_id"); v != "" {
		id, _ := strconv.ParseUint(v, 10, 64)
		uid := uint(id)
		filter.UserID = &uid
	}
	if v := c.QueryParam("status"); v != "" {
		filter.Status = entity.AttendanceStatus(v)
	}
	if v := c.QueryParam("incomplete"); v != "" {
		filter.Incomplete = v // "checkin" | "checkout" | "any"
	}
	if v := c.QueryParam("date_from"); v != "" {
		t, err := utils.ParseDateHCM( v)
		if err == nil {
			filter.DateFrom = &t
		}
	}
	if v := c.QueryParam("date_to"); v != "" {
		t, err := utils.ParseDateHCM( v)
		if err == nil {
			filter.DateTo = &t
		}
	}

	// Manager chỉ xem được data của chi nhánh mình
	if middleware.GetRole(c) == "manager" {
		branchID := middleware.GetBranchID(c)
		filter.BranchID = branchID
	}

	logs, total, err := h.attendanceUsecase.GetList(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, logs, total, pagination.Page, pagination.Limit)
}

// GetByID godoc
// @Summary Lấy chi tiết một bản ghi chấm công
// @Tags Admin - Attendance
// @Security BearerAuth
// @Produce json
// @Param id path int true "Attendance ID"
// @Success 200 {object} response.Response{data=entity.AttendanceLog}
// @Router /admin/attendance/{id} [get]
func (h *AttendanceHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	log, err := h.attendanceUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, log)
}

// GetSummary godoc
// @Summary Lấy tổng hợp thống kê chấm công của nhân viên
// @Tags Admin - Attendance
// @Security BearerAuth
// @Produce json
// @Param user_id path int true "User ID"
// @Param date_from query string true "Từ ngày (YYYY-MM-DD)"
// @Param date_to query string true "Đến ngày (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=repository.AttendanceSummary}
// @Router /admin/attendance/summary/{user_id} [get]
func (h *AttendanceHandler) GetSummary(c echo.Context) error {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	from, err := utils.ParseDateHCM( c.QueryParam("date_from"))
	if err != nil {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"date_from": "date_from không hợp lệ (định dạng YYYY-MM-DD)",
		}))
	}
	to, err := utils.ParseDateHCM( c.QueryParam("date_to"))
	if err != nil {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"date_to": "date_to không hợp lệ (định dạng YYYY-MM-DD)",
		}))
	}

	summary, err := h.attendanceUsecase.GetSummary(c.Request().Context(), uint(userID), from, to)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, summary)
}
