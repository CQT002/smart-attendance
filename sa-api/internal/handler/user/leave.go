package user

import (
	"strconv"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/labstack/echo/v4"
)

// LeaveHandler xử lý API nghỉ phép dành cho nhân viên
type LeaveHandler struct {
	leaveUsecase usecase.LeaveUsecase
}

// NewLeaveHandler tạo instance LeaveHandler
func NewLeaveHandler(leaveUsecase usecase.LeaveUsecase) *LeaveHandler {
	return &LeaveHandler{leaveUsecase: leaveUsecase}
}

// Create godoc
// @Summary Tạo yêu cầu nghỉ phép
// @Description Nhân viên đăng ký nghỉ phép cho ngày quá khứ, hiện tại hoặc tương lai
// @Tags Leave
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.CreateLeaveRequest true "Thông tin yêu cầu nghỉ phép"
// @Success 201 {object} response.Response{data=entity.LeaveRequest}
// @Router /attendance/leaves [post]
func (h *LeaveHandler) Create(c echo.Context) error {
	var req usecase.CreateLeaveRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	req.UserID = getUserIDFromContext(c)

	result, err := h.leaveUsecase.Create(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, result)
}

// GetMyList godoc
// @Summary Lấy danh sách yêu cầu nghỉ phép của bản thân
// @Tags Leave
// @Security BearerAuth
// @Produce json
// @Param status query string false "Lọc theo trạng thái (pending/approved/rejected)"
// @Param page query int false "Trang" default(1)
// @Param limit query int false "Số bản ghi mỗi trang" default(20)
// @Success 200 {object} response.Response{data=[]entity.LeaveRequest}
// @Router /attendance/leaves [get]
func (h *LeaveHandler) GetMyList(c echo.Context) error {
	userID := getUserIDFromContext(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	status := entity.LeaveStatus(c.QueryParam("status"))

	leaves, total, err := h.leaveUsecase.GetMyList(c.Request().Context(), userID, status, page, limit)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, leaves, total, page, limit)
}

// GetByID godoc
// @Summary Lấy chi tiết yêu cầu nghỉ phép
// @Tags Leave
// @Security BearerAuth
// @Produce json
// @Param id path int true "Leave ID"
// @Success 200 {object} response.Response{data=entity.LeaveRequest}
// @Router /attendance/leaves/{id} [get]
func (h *LeaveHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	leave, err := h.leaveUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	// Employee chỉ xem được yêu cầu của mình
	userID := getUserIDFromContext(c)
	if leave.UserID != userID {
		return response.Error(c, apperrors.ErrForbidden)
	}

	return response.OK(c, leave)
}
