package user

import (
	"strconv"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/labstack/echo/v4"
)

// OvertimeHandler xử lý API tăng ca dành cho nhân viên
type OvertimeHandler struct {
	overtimeUsecase usecase.OvertimeUsecase
}

// NewOvertimeHandler tạo instance OvertimeHandler
func NewOvertimeHandler(overtimeUsecase usecase.OvertimeUsecase) *OvertimeHandler {
	return &OvertimeHandler{overtimeUsecase: overtimeUsecase}
}

// CheckIn godoc
// @Summary Check-in tăng ca
// @Description Chỉ cho phép check-in OT sau 17:00. Giờ tính OT từ 18:00 đến 22:00.
// @Tags Overtime
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=usecase.OvertimeCheckInResponse}
// @Failure 400 {object} response.Response "Check-in trước 17:00"
// @Router /attendance/overtime/check-in [post]
func (h *OvertimeHandler) CheckIn(c echo.Context) error {
	req := usecase.OvertimeCheckInRequest{
		UserID: getUserIDFromContext(c),
	}

	result, err := h.overtimeUsecase.CheckIn(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Check-in tăng ca thành công", result)
}

// CheckOut godoc
// @Summary Check-out tăng ca
// @Tags Overtime
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.OvertimeCheckOutRequest true "Thông tin check-out OT"
// @Success 200 {object} response.Response{data=usecase.OvertimeCheckOutResponse}
// @Router /attendance/overtime/check-out [post]
func (h *OvertimeHandler) CheckOut(c echo.Context) error {
	var req usecase.OvertimeCheckOutRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	req.UserID = getUserIDFromContext(c)

	result, err := h.overtimeUsecase.CheckOut(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Check-out tăng ca thành công", result)
}

// GetMyToday godoc
// @Summary Lấy trạng thái tăng ca hôm nay
// @Tags Overtime
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=entity.OvertimeRequest}
// @Router /attendance/overtime/today [get]
func (h *OvertimeHandler) GetMyToday(c echo.Context) error {
	userID := getUserIDFromContext(c)
	ot, err := h.overtimeUsecase.GetMyToday(c.Request().Context(), userID)
	if err != nil {
		return response.Error(c, err)
	}
	return response.OK(c, ot)
}

// GetMyList godoc
// @Summary Lấy lịch sử tăng ca của bản thân
// @Tags Overtime
// @Security BearerAuth
// @Produce json
// @Param status query string false "Lọc theo trạng thái (pending/approved/rejected)"
// @Param page query int false "Trang" default(1)
// @Param limit query int false "Số bản ghi mỗi trang" default(20)
// @Success 200 {object} response.Response{data=[]entity.OvertimeRequest}
// @Router /attendance/overtime [get]
func (h *OvertimeHandler) GetMyList(c echo.Context) error {
	userID := getUserIDFromContext(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	status := entity.OvertimeStatus(c.QueryParam("status"))

	overtimes, total, err := h.overtimeUsecase.GetMyList(c.Request().Context(), userID, status, page, limit)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, overtimes, total, page, limit)
}

// GetByID godoc
// @Summary Lấy chi tiết yêu cầu tăng ca
// @Tags Overtime
// @Security BearerAuth
// @Produce json
// @Param id path int true "Overtime ID"
// @Success 200 {object} response.Response{data=entity.OvertimeRequest}
// @Router /attendance/overtime/{id} [get]
func (h *OvertimeHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	ot, err := h.overtimeUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	// Employee chỉ xem được yêu cầu của mình
	userID := getUserIDFromContext(c)
	if ot.UserID != userID {
		return response.Error(c, apperrors.ErrForbidden)
	}

	return response.OK(c, ot)
}
