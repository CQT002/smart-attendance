package user

import (
	"strconv"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/labstack/echo/v4"
)

// AttendanceHandler xử lý các API chấm công dành cho nhân viên
type AttendanceHandler struct {
	attendanceUsecase usecase.AttendanceUsecase
}

// NewAttendanceHandler tạo instance AttendanceHandler
func NewAttendanceHandler(attendanceUsecase usecase.AttendanceUsecase) *AttendanceHandler {
	return &AttendanceHandler{attendanceUsecase: attendanceUsecase}
}

// CheckIn godoc
// @Summary Check-in chấm công
// @Description Xác thực vị trí qua WiFi hoặc GPS, kiểm tra chống gian lận trước khi ghi nhận
// @Tags Attendance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.CheckInRequest true "Thông tin check-in"
// @Success 200 {object} response.Response{data=entity.AttendanceLog}
// @Failure 403 {object} response.Response "Vị trí không hợp lệ hoặc phát hiện gian lận"
// @Router /attendance/check-in [post]
func (h *AttendanceHandler) CheckIn(c echo.Context) error {
	var req usecase.CheckInRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	// Gắn userID từ JWT token vào request
	req.UserID = getUserIDFromContext(c)
	req.IPAddress = c.RealIP()

	result, err := h.attendanceUsecase.CheckIn(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Check-in thành công", result)
}

// CheckOut godoc
// @Summary Check-out kết thúc ca làm
// @Tags Attendance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.CheckOutRequest true "Thông tin check-out"
// @Success 200 {object} response.Response{data=entity.AttendanceLog}
// @Router /attendance/check-out [post]
func (h *AttendanceHandler) CheckOut(c echo.Context) error {
	var req usecase.CheckOutRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	req.UserID = getUserIDFromContext(c)
	req.IPAddress = c.RealIP()

	if req.AttendanceID == 0 {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"attendance_id": "attendance_id không được để trống",
		}))
	}

	result, err := h.attendanceUsecase.CheckOut(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Check-out thành công", result)
}

// GetMyToday godoc
// @Summary Lấy trạng thái chấm công hôm nay của bản thân
// @Tags Attendance
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=entity.AttendanceLog}
// @Router /attendance/today [get]
func (h *AttendanceHandler) GetMyToday(c echo.Context) error {
	userID := getUserIDFromContext(c)
	log, err := h.attendanceUsecase.GetMyToday(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, log)
}

// GetMyHistory godoc
// @Summary Lấy lịch sử chấm công của bản thân
// @Tags Attendance
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Ngày bắt đầu (YYYY-MM-DD)"
// @Param date_to query string false "Ngày kết thúc (YYYY-MM-DD)"
// @Param page query int false "Trang" default(1)
// @Param limit query int false "Số bản ghi mỗi trang" default(20)
// @Success 200 {object} response.Response{data=[]entity.AttendanceLog}
// @Router /attendance/history [get]
func (h *AttendanceHandler) GetMyHistory(c echo.Context) error {
	userID := getUserIDFromContext(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	filter := repository.AttendanceFilter{
		UserID: &userID,
		Page:   page,
		Limit:  limit,
	}

	if dateFrom := c.QueryParam("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateTo := c.QueryParam("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			filter.DateTo = &endOfDay
		}
	}

	records, total, err := h.attendanceUsecase.GetList(c.Request().Context(), filter)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, records, total, page, limit)
}
