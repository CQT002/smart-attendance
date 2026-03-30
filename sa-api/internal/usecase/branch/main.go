package branch

import (
	"context"
	"fmt"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/internal/infrastructure/cache"
)

type branchUsecase struct {
	branchRepo repository.BranchRepository
	cache      cache.Cache
}

// NewBranchUsecase tạo instance BranchUsecase
func NewBranchUsecase(branchRepo repository.BranchRepository, cache cache.Cache) usecase.BranchUsecase {
	return &branchUsecase{branchRepo: branchRepo, cache: cache}
}

func (u *branchUsecase) Create(ctx context.Context, req usecase.CreateBranchRequest) (*entity.Branch, error) {
	branch := &entity.Branch{
		Code:     req.Code,
		Name:     req.Name,
		Address:  req.Address,
		Phone:    req.Phone,
		Email:    req.Email,
		IsActive: true,
	}

	if err := u.branchRepo.Create(ctx, branch); err != nil {
		return nil, err
	}

	// Xóa cache danh sách chi nhánh
	u.cache.DeletePattern(ctx, cache.KeyPrefixBranch+"*")

	return branch, nil
}

func (u *branchUsecase) Update(ctx context.Context, id uint, req usecase.UpdateBranchRequest) (*entity.Branch, error) {
	branch, err := u.branchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	branch.Name = req.Name
	branch.Address = req.Address
	branch.Phone = req.Phone
	branch.Email = req.Email

	if err := u.branchRepo.Update(ctx, branch); err != nil {
		return nil, err
	}

	// Xóa cache
	u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixBranch, fmt.Sprintf("%d", id)))
	u.cache.DeletePattern(ctx, cache.KeyPrefixBranch+"list:*")

	return branch, nil
}

func (u *branchUsecase) Delete(ctx context.Context, id uint) error {
	if err := u.branchRepo.Delete(ctx, id); err != nil {
		return err
	}
	u.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixBranch, fmt.Sprintf("%d", id)))
	u.cache.DeletePattern(ctx, cache.KeyPrefixBranch+"list:*")
	return nil
}

// GetByID lấy thông tin chi nhánh, ưu tiên cache
func (u *branchUsecase) GetByID(ctx context.Context, id uint) (*entity.Branch, error) {
	cacheKey := cache.BuildKey(cache.KeyPrefixBranch, fmt.Sprintf("%d", id))

	var cached entity.Branch
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	branch, err := u.branchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache 30 phút vì branch thay đổi ít thường xuyên
	u.cache.Set(ctx, cacheKey, branch, 30*time.Minute)
	return branch, nil
}

func (u *branchUsecase) GetList(ctx context.Context, filter repository.BranchFilter) ([]*entity.Branch, int64, error) {
	return u.branchRepo.FindAll(ctx, filter)
}

// GetActive lấy danh sách chi nhánh active, cache 15 phút
func (u *branchUsecase) GetActive(ctx context.Context) ([]*entity.Branch, error) {
	cacheKey := cache.KeyPrefixBranch + "active"

	var cached []*entity.Branch
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	branches, err := u.branchRepo.FindActive(ctx)
	if err != nil {
		return nil, err
	}

	u.cache.Set(ctx, cacheKey, branches, 15*time.Minute)
	return branches, nil
}
