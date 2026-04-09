package admin

import (
	"log/slog"
	"strconv"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/labstack/echo/v4"
)

// ShiftHandler xử lý các API quản lý ca làm việc của chi nhánh
type ShiftHandler struct {
	shiftRepo repository.ShiftRepository
}

// NewShiftHandler tạo instance ShiftHandler
func NewShiftHandler(shiftRepo repository.ShiftRepository) *ShiftHandler {
	return &ShiftHandler{shiftRepo: shiftRepo}
}

// CreateShiftRequest yêu cầu tạo ca làm việc
type CreateShiftRequest struct {
	Name           string  `json:"name"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	LateAfter      int     `json:"late_after"`
	EarlyBefore    int     `json:"early_before"`
	WorkHours      float64 `json:"work_hours"`
	MorningEnd     string  `json:"morning_end"`
	AfternoonStart string  `json:"afternoon_start"`
	RegularEndDay  *int    `json:"regular_end_day"`
	RegularEndTime string  `json:"regular_end_time"`
	OTMinCheckInHour int   `json:"ot_min_checkin_hour"`
	OTStartHour    int     `json:"ot_start_hour"`
	OTEndHour      int     `json:"ot_end_hour"`
	IsDefault      bool    `json:"is_default"`
}

// UpdateShiftRequest yêu cầu cập nhật ca làm việc
type UpdateShiftRequest struct {
	Name           *string  `json:"name"`
	StartTime      *string  `json:"start_time"`
	EndTime        *string  `json:"end_time"`
	LateAfter      *int     `json:"late_after"`
	EarlyBefore    *int     `json:"early_before"`
	WorkHours      *float64 `json:"work_hours"`
	MorningEnd     *string  `json:"morning_end"`
	AfternoonStart *string  `json:"afternoon_start"`
	RegularEndDay  *int     `json:"regular_end_day"`
	RegularEndTime *string  `json:"regular_end_time"`
	OTMinCheckInHour *int   `json:"ot_min_checkin_hour"`
	OTStartHour    *int     `json:"ot_start_hour"`
	OTEndHour      *int     `json:"ot_end_hour"`
	IsDefault      *bool    `json:"is_default"`
	IsActive       *bool    `json:"is_active"`
}

// GetByBranch godoc
// @Summary Lấy danh sách ca làm việc của chi nhánh
// @Tags Admin - Shift
// @Security BearerAuth
// @Produce json
// @Param branch_id path int true "Branch ID"
// @Success 200 {object} response.Response{data=[]entity.Shift}
// @Router /admin/branches/{branch_id}/shifts [get]
func (h *ShiftHandler) GetByBranch(c echo.Context) error {
	branchID, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	shifts, err := h.shiftRepo.FindByBranch(c.Request().Context(), uint(branchID))
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, shifts)
}

// Create godoc
// @Summary Tạo ca làm việc cho chi nhánh (Admin)
// @Tags Admin - Shift
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param branch_id path int true "Branch ID"
// @Param body body CreateShiftRequest true "Thông tin ca làm việc"
// @Success 201 {object} response.Response{data=entity.Shift}
// @Router /admin/branches/{branch_id}/shifts [post]
func (h *ShiftHandler) Create(c echo.Context) error {
	branchID, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var req CreateShiftRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if req.Name == "" {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"name": "Tên ca không được để trống",
		}))
	}
	if req.StartTime == "" || req.EndTime == "" {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"start_time": "Giờ bắt đầu và kết thúc không được để trống",
		}))
	}

	shift := &entity.Shift{
		BranchID:       uint(branchID),
		Name:           req.Name,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		LateAfter:      req.LateAfter,
		EarlyBefore:    req.EarlyBefore,
		WorkHours:      req.WorkHours,
		MorningEnd:     req.MorningEnd,
		AfternoonStart: req.AfternoonStart,
		RegularEndDay:  6,
		RegularEndTime: "12:00",
		OTMinCheckInHour: req.OTMinCheckInHour,
		OTStartHour:    req.OTStartHour,
		OTEndHour:      req.OTEndHour,
		IsDefault:      req.IsDefault,
		IsActive:       true,
	}

	if req.RegularEndDay != nil {
		shift.RegularEndDay = *req.RegularEndDay
	}
	if req.RegularEndTime != "" {
		shift.RegularEndTime = req.RegularEndTime
	}

	// Defaults
	if shift.LateAfter == 0 {
		shift.LateAfter = 15
	}
	if shift.EarlyBefore == 0 {
		shift.EarlyBefore = 15
	}
	if shift.WorkHours == 0 {
		shift.WorkHours = 8
	}
	if shift.MorningEnd == "" {
		shift.MorningEnd = "12:00"
	}
	if shift.AfternoonStart == "" {
		shift.AfternoonStart = "13:00"
	}
	if shift.OTMinCheckInHour == 0 {
		shift.OTMinCheckInHour = 17
	}
	if shift.OTStartHour == 0 {
		shift.OTStartHour = 18
	}
	if shift.OTEndHour == 0 {
		shift.OTEndHour = 22
	}

	if err := h.shiftRepo.Create(c.Request().Context(), shift); err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, shift)
}

// Update godoc
// @Summary Cập nhật ca làm việc (Admin)
// @Tags Admin - Shift
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param branch_id path int true "Branch ID"
// @Param id path int true "Shift ID"
// @Param body body UpdateShiftRequest true "Thông tin cập nhật"
// @Success 200 {object} response.Response{data=entity.Shift}
// @Router /admin/branches/{branch_id}/shifts/{id} [put]
func (h *ShiftHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	shift, err := h.shiftRepo.FindByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	var req UpdateShiftRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	// Partial update
	if req.Name != nil {
		shift.Name = *req.Name
	}
	if req.StartTime != nil {
		shift.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		shift.EndTime = *req.EndTime
	}
	if req.LateAfter != nil {
		shift.LateAfter = *req.LateAfter
	}
	if req.EarlyBefore != nil {
		shift.EarlyBefore = *req.EarlyBefore
	}
	if req.WorkHours != nil {
		shift.WorkHours = *req.WorkHours
	}
	if req.MorningEnd != nil {
		shift.MorningEnd = *req.MorningEnd
	}
	if req.AfternoonStart != nil {
		shift.AfternoonStart = *req.AfternoonStart
	}
	if req.RegularEndDay != nil {
		shift.RegularEndDay = *req.RegularEndDay
	}
	if req.RegularEndTime != nil {
		shift.RegularEndTime = *req.RegularEndTime
	}
	if req.OTMinCheckInHour != nil {
		shift.OTMinCheckInHour = *req.OTMinCheckInHour
	}
	if req.OTStartHour != nil {
		shift.OTStartHour = *req.OTStartHour
	}
	if req.OTEndHour != nil {
		shift.OTEndHour = *req.OTEndHour
	}
	if req.IsDefault != nil {
		shift.IsDefault = *req.IsDefault
	}
	if req.IsActive != nil {
		shift.IsActive = *req.IsActive
	}

	if err := h.shiftRepo.Update(c.Request().Context(), shift); err != nil {
		slog.Error("shift update failed", "shift_id", shift.ID, "error", err)
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Cập nhật ca làm việc thành công", shift)
}

// Delete godoc
// @Summary Xóa ca làm việc (Admin)
// @Tags Admin - Shift
// @Security BearerAuth
// @Produce json
// @Param branch_id path int true "Branch ID"
// @Param id path int true "Shift ID"
// @Success 204
// @Router /admin/branches/{branch_id}/shifts/{id} [delete]
func (h *ShiftHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if err := h.shiftRepo.Delete(c.Request().Context(), uint(id)); err != nil {
		return response.Error(c, err)
	}

	return response.NoContent(c)
}
