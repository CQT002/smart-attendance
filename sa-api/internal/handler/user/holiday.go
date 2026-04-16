package user

import (
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	ucHoliday "github.com/hdbank/smart-attendance/internal/usecase/holiday"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"github.com/labstack/echo/v4"
)

// HolidayHandler xử lý API ngày lễ cho employee (read-only).
type HolidayHandler struct {
	holidayUsecase usecase.HolidayUsecase
	calculator     *ucHoliday.Calculator
	userRepo       repository.UserRepository
}

// NewHolidayHandler tạo instance
func NewHolidayHandler(
	holidayUC usecase.HolidayUsecase,
	calculator *ucHoliday.Calculator,
	userRepo repository.UserRepository,
) *HolidayHandler {
	return &HolidayHandler{
		holidayUsecase: holidayUC,
		calculator:     calculator,
		userRepo:       userRepo,
	}
}

// GetCalendar godoc
// @Summary Lấy danh sách ngày lễ cho mobile calendar
// @Description Mặc định trả về ngày lễ trong năm hiện tại nếu không truyền query
// @Tags Holiday
// @Security BearerAuth
// @Produce json
// @Param date_from query string false "Từ ngày (YYYY-MM-DD)"
// @Param date_to query string false "Đến ngày (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=[]entity.Holiday}
// @Router /holidays/calendar [get]
func (h *HolidayHandler) GetCalendar(c echo.Context) error {
	from, to, err := parseRangeOrDefaultYear(c)
	if err != nil {
		return response.Error(c, err)
	}

	holidays, err := h.holidayUsecase.GetCalendar(c.Request().Context(), from, to)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, holidays)
}

// GetSummary godoc
// @Summary Lấy báo cáo công kèm thông tin ngày lễ
// @Description Highlight ngày lễ + hiển thị hệ số lương. Trả về detail per-day + aggregate.
// @Tags Attendance
// @Security BearerAuth
// @Produce json
// @Param date_from query string true "Từ ngày (YYYY-MM-DD)"
// @Param date_to query string true "Đến ngày (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=holiday.AttendanceCalculationResult}
// @Router /attendance/summary [get]
func (h *HolidayHandler) GetSummary(c echo.Context) error {
	userID := getUserIDFromContext(c)

	from, err := utils.ParseDateHCM(c.QueryParam("date_from"))
	if err != nil {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"date_from": "date_from không hợp lệ (YYYY-MM-DD)",
		}))
	}
	to, err := utils.ParseDateHCM(c.QueryParam("date_to"))
	if err != nil {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"date_to": "date_to không hợp lệ (YYYY-MM-DD)",
		}))
	}
	if to.Before(from) {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"date_to": "date_to phải sau date_from",
		}))
	}

	u, err := h.userRepo.FindByID(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, err)
	}
	var branchID uint
	if u.BranchID != nil {
		branchID = *u.BranchID
	}

	result, err := h.calculator.CalculateAttendanceLog(c.Request().Context(), userID, branchID, from, to)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, result)
}

// parseRangeOrDefaultYear fallback năm hiện tại nếu không có query.
func parseRangeOrDefaultYear(c echo.Context) (time.Time, time.Time, error) {
	fromStr := c.QueryParam("date_from")
	toStr := c.QueryParam("date_to")
	if fromStr == "" && toStr == "" {
		now := utils.Now()
		from := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, utils.HCM)
		to := time.Date(now.Year(), time.December, 31, 0, 0, 0, 0, utils.HCM)
		return from, to, nil
	}
	from, err := utils.ParseDateHCM(fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, apperrors.NewValidationError(map[string]string{
			"date_from": "date_from không hợp lệ (YYYY-MM-DD)",
		})
	}
	to, err := utils.ParseDateHCM(toStr)
	if err != nil {
		return time.Time{}, time.Time{}, apperrors.NewValidationError(map[string]string{
			"date_to": "date_to không hợp lệ (YYYY-MM-DD)",
		})
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, apperrors.NewValidationError(map[string]string{
			"date_to": "date_to phải sau date_from",
		})
	}
	return from, to, nil
}
