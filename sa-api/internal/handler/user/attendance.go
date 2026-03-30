package user

import (
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
