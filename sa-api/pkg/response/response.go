package response

import (
	"net/http"

	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/labstack/echo/v4"
)

// Response cấu trúc response chuẩn cho tất cả API
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// ErrorDetail chi tiết lỗi trả về client
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Meta metadata cho danh sách có phân trang
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// OK trả về response thành công
func OK(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// OKWithMessage trả về response thành công kèm thông báo
func OKWithMessage(c echo.Context, message string, data any) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created trả về response tạo mới thành công
func Created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// Paginated trả về response danh sách có phân trang
func Paginated(c echo.Context, data any, total int64, page, limit int) error {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error trả về response lỗi chuẩn hóa
func Error(c echo.Context, err error) error {
	if appErr, ok := apperrors.IsAppError(err); ok {
		detail := &ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
		// Kiểm tra có phải validation error không
		if ve, ok := err.(*apperrors.ValidationError); ok {
			detail.Fields = ve.Fields
		}
		return c.JSON(appErr.HTTPStatus, Response{
			Success: false,
			Error:   detail,
		})
	}

	// Lỗi không xác định - trả về 500
	return c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorDetail{
			Code:    "INTERNAL_ERROR",
			Message: "Lỗi hệ thống, vui lòng thử lại sau",
		},
	})
}

// NoContent trả về 204 khi xóa thành công
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
