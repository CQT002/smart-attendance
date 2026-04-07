package user

import (
	"strconv"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/labstack/echo/v4"
)

// CorrectionHandler xử lý API chấm công bù dành cho nhân viên
type CorrectionHandler struct {
	correctionUsecase usecase.CorrectionUsecase
}

// NewCorrectionHandler tạo instance CorrectionHandler
func NewCorrectionHandler(correctionUsecase usecase.CorrectionUsecase) *CorrectionHandler {
	return &CorrectionHandler{correctionUsecase: correctionUsecase}
}

// Create godoc
// @Summary Tạo yêu cầu chấm công bù
// @Description Nhân viên đăng ký bù công cho ngày bị trễ/về sớm (tối đa 4 lần/tháng)
// @Tags Attendance Correction
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body usecase.CreateCorrectionRequest true "Thông tin yêu cầu bù công"
// @Success 201 {object} response.Response{data=entity.AttendanceCorrection}
// @Failure 400 {object} response.Response "Hạn mức/trạng thái không hợp lệ"
// @Router /attendance/corrections [post]
func (h *CorrectionHandler) Create(c echo.Context) error {
	var req usecase.CreateCorrectionRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	req.UserID = getUserIDFromContext(c)

	result, err := h.correctionUsecase.Create(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, result)
}

// GetMyList godoc
// @Summary Lấy danh sách yêu cầu chấm công bù của bản thân
// @Tags Attendance Correction
// @Security BearerAuth
// @Produce json
// @Param status query string false "Lọc theo trạng thái (pending/approved/rejected)"
// @Param page query int false "Trang" default(1)
// @Param limit query int false "Số bản ghi mỗi trang" default(20)
// @Success 200 {object} response.Response{data=[]entity.AttendanceCorrection}
// @Router /attendance/corrections [get]
func (h *CorrectionHandler) GetMyList(c echo.Context) error {
	userID := getUserIDFromContext(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	status := entity.CorrectionStatus(c.QueryParam("status"))

	corrections, total, err := h.correctionUsecase.GetMyList(c.Request().Context(), userID, status, page, limit)
	if err != nil {
		return response.Error(c, err)
	}

	return response.Paginated(c, corrections, total, page, limit)
}

// GetByID godoc
// @Summary Lấy chi tiết yêu cầu chấm công bù
// @Tags Attendance Correction
// @Security BearerAuth
// @Produce json
// @Param id path int true "Correction ID"
// @Success 200 {object} response.Response{data=entity.AttendanceCorrection}
// @Router /attendance/corrections/{id} [get]
func (h *CorrectionHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	correction, err := h.correctionUsecase.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	// Employee chỉ xem được yêu cầu của mình
	userID := getUserIDFromContext(c)
	if correction.UserID != userID {
		return response.Error(c, apperrors.ErrForbidden)
	}

	return response.OK(c, correction)
}
