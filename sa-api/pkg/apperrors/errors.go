package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError lỗi nghiệp vụ có HTTP status code
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New tạo AppError mới
func New(httpStatus int, code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap bọc lỗi gốc vào AppError
func Wrap(err error, httpStatus int, code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// IsAppError kiểm tra có phải AppError không
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// ===================== Lỗi chuẩn hóa theo domain =====================

// Authentication errors
var (
	ErrUnauthorized      = New(http.StatusUnauthorized, "UNAUTHORIZED", "Bạn chưa đăng nhập hoặc phiên đăng nhập đã hết hạn")
	ErrForbidden         = New(http.StatusForbidden, "FORBIDDEN", "Bạn không có quyền thực hiện thao tác này")
	ErrInvalidCredential = New(http.StatusUnauthorized, "INVALID_CREDENTIAL", "Email hoặc mật khẩu không đúng")
	ErrTokenExpired      = New(http.StatusUnauthorized, "TOKEN_EXPIRED", "Token đã hết hạn, vui lòng đăng nhập lại")
	ErrTokenInvalid      = New(http.StatusUnauthorized, "TOKEN_INVALID", "Token không hợp lệ")
)

// Resource errors
var (
	ErrNotFound        = New(http.StatusNotFound, "NOT_FOUND", "Không tìm thấy dữ liệu")
	ErrUserNotFound    = New(http.StatusNotFound, "USER_NOT_FOUND", "Không tìm thấy người dùng")
	ErrBranchNotFound  = New(http.StatusNotFound, "BRANCH_NOT_FOUND", "Không tìm thấy chi nhánh")
	ErrDuplicate       = New(http.StatusConflict, "DUPLICATE", "Dữ liệu đã tồn tại")
	ErrEmailDuplicate  = New(http.StatusConflict, "EMAIL_DUPLICATE", "Email đã được sử dụng")
	ErrCodeDuplicate   = New(http.StatusConflict, "CODE_DUPLICATE", "Mã đã tồn tại trong hệ thống")
)

// Attendance errors
var (
	ErrAlreadyCheckedIn    = New(http.StatusConflict, "ALREADY_CHECKED_IN", "Bạn đã check-in hôm nay")
	ErrNotCheckedIn        = New(http.StatusBadRequest, "NOT_CHECKED_IN", "Bạn chưa check-in")
	ErrAlreadyCheckedOut   = New(http.StatusConflict, "ALREADY_CHECKED_OUT", "Bạn đã check-out hôm nay")
	ErrLocationNotAllowed  = New(http.StatusForbidden, "LOCATION_NOT_ALLOWED", "Vị trí của bạn không nằm trong vùng cho phép chấm công")
	ErrWiFiNotAllowed      = New(http.StatusForbidden, "WIFI_NOT_ALLOWED", "Mạng WiFi hiện tại không được phép chấm công")
	ErrFakeGPSDetected     = New(http.StatusForbidden, "FAKE_GPS_DETECTED", "Phát hiện GPS giả, không thể chấm công")
	ErrVPNDetected         = New(http.StatusForbidden, "VPN_DETECTED", "Phát hiện kết nối VPN, không thể chấm công")
	ErrSuspiciousActivity  = New(http.StatusForbidden, "SUSPICIOUS_ACTIVITY", "Phát hiện hoạt động bất thường, tài khoản tạm khóa chấm công")
)

// Correction errors
var (
	ErrCorrectionLimitExceeded = New(http.StatusBadRequest, "CORRECTION_LIMIT_EXCEEDED", "Bạn đã sử dụng hết hạn mức chấm công bù trong tháng (tối đa 4 lần)")
	ErrCorrectionInvalidStatus = New(http.StatusBadRequest, "CORRECTION_INVALID_STATUS", "Chỉ được đăng ký bù công cho ngày đi trễ, về sớm hoặc đi trễ về sớm")
	ErrCorrectionAlreadyExists = New(http.StatusConflict, "CORRECTION_ALREADY_EXISTS", "Ngày này đã có yêu cầu chấm công bù")
	ErrCorrectionNotPending    = New(http.StatusBadRequest, "CORRECTION_NOT_PENDING", "Yêu cầu này đã được xử lý, không thể thay đổi")
	ErrCorrectionSelfApprove   = New(http.StatusForbidden, "CORRECTION_SELF_APPROVE", "Manager không được tự duyệt yêu cầu của chính mình")
)

// Leave errors
var (
	ErrLeaveAlreadyExists  = New(http.StatusConflict, "LEAVE_ALREADY_EXISTS", "Ngày này đã có yêu cầu nghỉ phép")
	ErrLeaveInvalidDate    = New(http.StatusBadRequest, "LEAVE_INVALID_DATE", "Ngày nghỉ phép không hợp lệ")
	ErrLeaveInvalidType    = New(http.StatusBadRequest, "LEAVE_INVALID_TYPE", "Loại nghỉ phép không hợp lệ cho ngày này")
	ErrLeaveNotPending     = New(http.StatusBadRequest, "LEAVE_NOT_PENDING", "Yêu cầu này đã được xử lý, không thể thay đổi")
	ErrLeaveSelfApprove    = New(http.StatusForbidden, "LEAVE_SELF_APPROVE", "Manager không được tự duyệt yêu cầu của chính mình")
	ErrLeaveInvalidStatus  = New(http.StatusBadRequest, "LEAVE_INVALID_STATUS", "Trạng thái ngày không cho phép xin nghỉ phép")
	ErrLeaveInsufficientBalance = New(http.StatusBadRequest, "LEAVE_INSUFFICIENT_BALANCE", "Số ngày phép không đủ")
)

// Overtime errors
var (
	ErrOvertimeCheckInTooEarly   = New(http.StatusBadRequest, "OT_CHECKIN_TOO_EARLY", "Chỉ được check-in tăng ca sau 17:00")
	ErrOvertimeAlreadyCheckedIn  = New(http.StatusConflict, "OT_ALREADY_CHECKED_IN", "Bạn đã check-in tăng ca hôm nay")
	ErrOvertimeNotCheckedIn      = New(http.StatusBadRequest, "OT_NOT_CHECKED_IN", "Bạn chưa check-in tăng ca")
	ErrOvertimeAlreadyCheckedOut = New(http.StatusConflict, "OT_ALREADY_CHECKED_OUT", "Bạn đã check-out tăng ca hôm nay")
	ErrOvertimeNotPending        = New(http.StatusBadRequest, "OT_NOT_PENDING", "Yêu cầu tăng ca đã được xử lý, không thể thay đổi")
	ErrOvertimeSelfApprove       = New(http.StatusForbidden, "OT_SELF_APPROVE", "Manager không được tự duyệt yêu cầu tăng ca của chính mình")
	ErrOvertimeNotCompleted      = New(http.StatusBadRequest, "OT_NOT_COMPLETED", "Yêu cầu tăng ca chưa hoàn tất (thiếu check-out)")
	ErrOvertimeAlreadyExists     = New(http.StatusConflict, "OT_ALREADY_EXISTS", "Ngày này đã có yêu cầu tăng ca")
	ErrOvertimeCheckOutTooEarly  = New(http.StatusBadRequest, "OT_CHECKOUT_TOO_EARLY", "Chỉ được check-out tăng ca sau 18:00")
)

// Validation errors
var (
	ErrValidation      = New(http.StatusBadRequest, "VALIDATION_ERROR", "Dữ liệu không hợp lệ")
	ErrInvalidPassword = New(http.StatusBadRequest, "INVALID_PASSWORD", "Mật khẩu cũ không đúng")
)

// Server errors
var (
	ErrInternal = New(http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống, vui lòng thử lại sau")
)

// ValidationError lỗi validate với chi tiết từng field
type ValidationError struct {
	*AppError
	Fields map[string]string `json:"fields"`
}

// NewValidationError tạo lỗi validation với danh sách field lỗi
func NewValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{
		AppError: ErrValidation,
		Fields:   fields,
	}
}
