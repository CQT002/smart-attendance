package usecase

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
)

// CreateCorrectionRequest yêu cầu tạo chấm công bù từ nhân viên
type CreateCorrectionRequest struct {
	UserID            uint                    `json:"-"`                      // Từ JWT, không trust client
	CorrectionType    entity.CorrectionType   `json:"correction_type"`        // "attendance" hoặc "overtime"
	AttendanceLogID   uint                    `json:"attendance_log_id"`      // ID bản ghi chấm công gốc (dùng cho attendance)
	OvertimeRequestID uint                    `json:"overtime_request_id"`    // ID OT request gốc (dùng cho overtime)
	Description       string                  `json:"description"`            // Lý do xin bù công
}

// ProcessCorrectionRequest yêu cầu duyệt/từ chối từ Manager
type ProcessCorrectionRequest struct {
	CorrectionID  uint                     `json:"-"`             // Từ path param
	ProcessedByID uint                     `json:"-"`             // Từ JWT
	Status        entity.CorrectionStatus  `json:"status"`        // approved hoặc rejected
	ManagerNote   string                   `json:"manager_note"`  // Ghi chú từ manager
}

// CorrectionUsecase định nghĩa business logic cho chấm công bù
type CorrectionUsecase interface {
	// Create tạo yêu cầu chấm công bù (employee)
	// Validate: hạn mức 4 lần/tháng, trạng thái ngày gốc, chưa có yêu cầu trùng
	Create(ctx context.Context, req CreateCorrectionRequest) (*entity.AttendanceCorrection, error)

	// Process duyệt hoặc từ chối yêu cầu (manager)
	// Khi approved: cập nhật AttendanceLog gốc thành VALIDATED trong transaction
	Process(ctx context.Context, req ProcessCorrectionRequest) (*entity.AttendanceCorrection, error)

	// GetByID lấy chi tiết yêu cầu
	GetByID(ctx context.Context, id uint) (*entity.AttendanceCorrection, error)

	// GetList lấy danh sách yêu cầu có phân trang và lọc
	GetList(ctx context.Context, filter repository.CorrectionFilter) ([]*entity.AttendanceCorrection, int64, error)

	// GetMyList lấy danh sách yêu cầu của employee (filter theo userID)
	GetMyList(ctx context.Context, userID uint, status entity.CorrectionStatus, page, limit int) ([]*entity.AttendanceCorrection, int64, error)

	// BatchApprove duyệt tất cả yêu cầu PENDING (theo branch nếu có)
	BatchApprove(ctx context.Context, processedByID uint, branchID *uint) (int64, error)

	// AutoRejectExpired tự động reject yêu cầu PENDING của tháng cũ (cron job)
	AutoRejectExpired(ctx context.Context) (int64, error)
}
