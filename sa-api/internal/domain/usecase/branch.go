package usecase

import (
	"context"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
)

// CreateBranchRequest yêu cầu tạo chi nhánh
type CreateBranchRequest struct {
	Code      string   `json:"code"`
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	City      string   `json:"city"`
	Province  string   `json:"province"`
	Phone     string   `json:"phone"`
	Email     string   `json:"email"`
	// GPS config — lưu vào gps_configs table, không lưu trên branches
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	GPSRadius *float64 `json:"gps_radius"`
}

// UpdateBranchRequest yêu cầu cập nhật chi nhánh
type UpdateBranchRequest struct {
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	City      string   `json:"city"`
	Province  string   `json:"province"`
	Phone     string   `json:"phone"`
	Email     string   `json:"email"`
	// GPS config — lưu vào gps_configs table, không lưu trên branches
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	GPSRadius *float64 `json:"gps_radius"`
}

// BranchUsecase định nghĩa business logic cho chi nhánh
type BranchUsecase interface {
	// Create tạo mới chi nhánh
	Create(ctx context.Context, req CreateBranchRequest) (*entity.Branch, error)

	// Update cập nhật thông tin chi nhánh
	Update(ctx context.Context, id uint, req UpdateBranchRequest) (*entity.Branch, error)

	// Delete vô hiệu hóa chi nhánh
	Delete(ctx context.Context, id uint) error

	// GetByID lấy thông tin chi nhánh theo ID
	GetByID(ctx context.Context, id uint) (*entity.Branch, error)

	// GetList lấy danh sách chi nhánh với phân trang và lọc
	GetList(ctx context.Context, filter repository.BranchFilter) ([]*entity.Branch, int64, error)

	// GetActive lấy danh sách chi nhánh đang hoạt động
	GetActive(ctx context.Context) ([]*entity.Branch, error)
}
