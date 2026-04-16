package holiday

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/internal/infrastructure/cache"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/utils"
)

const (
	cacheTTL          = 24 * time.Hour
	cacheKeyPrefix    = "holiday:"
	cacheKeyDate      = cacheKeyPrefix + "date:"     // holiday:date:2026-04-30
	cacheKeyRange     = cacheKeyPrefix + "range:"    // holiday:range:2026-01-01:2026-12-31
	cacheInvalidateAll = cacheKeyPrefix + "*"
)

type holidayUsecase struct {
	repo  repository.HolidayRepository
	cache cache.Cache
}

// NewHolidayUsecase tạo instance HolidayUsecase
func NewHolidayUsecase(repo repository.HolidayRepository, cache cache.Cache) usecase.HolidayUsecase {
	return &holidayUsecase{repo: repo, cache: cache}
}

// Create tạo ngày lễ mới. Tự derive year từ date, auto-fill coefficient theo type nếu = 0.
func (u *holidayUsecase) Create(ctx context.Context, req usecase.CreateHolidayRequest) (*entity.Holiday, error) {
	date, err := utils.ParseDateHCM(req.Date)
	if err != nil {
		return nil, apperrors.ErrHolidayInvalidDate
	}

	hType := req.Type
	if hType == "" {
		hType = entity.HolidayTypeNational
	}
	if hType != entity.HolidayTypeNational && hType != entity.HolidayTypeCompany {
		return nil, apperrors.ErrHolidayInvalidType
	}

	coef := req.Coefficient
	if coef == 0 {
		coef = entity.DefaultCoefficientFor(hType)
	}
	if coef <= 0 || coef > 10 {
		return nil, apperrors.ErrHolidayInvalidCoefficient
	}

	if req.Name == "" {
		return nil, apperrors.NewValidationError(map[string]string{"name": "Tên ngày lễ không được để trống"})
	}

	// Validate compensated
	var compensateFor *time.Time
	if req.IsCompensated {
		if req.CompensateFor == "" {
			return nil, apperrors.ErrHolidayCompensateMissing
		}
		cf, err := utils.ParseDateHCM(req.CompensateFor)
		if err != nil {
			return nil, apperrors.ErrHolidayCompensateMissing
		}
		compensateFor = &cf
	}

	// Uniqueness
	exists, err := u.repo.ExistsByDate(ctx, date, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.ErrHolidayAlreadyExists
	}

	h := &entity.Holiday{
		Name:          req.Name,
		Date:          date,
		Year:          date.Year(),
		Coefficient:   coef,
		Type:          hType,
		IsCompensated: req.IsCompensated,
		CompensateFor: compensateFor,
		Description:   req.Description,
	}
	if req.CreatedByID > 0 {
		id := req.CreatedByID
		h.CreatedByID = &id
	}

	if err := u.repo.Create(ctx, h); err != nil {
		return nil, err
	}

	u.invalidateCache(ctx)
	slog.Info("holiday created",
		"id", h.ID, "date", h.Date.Format("2006-01-02"),
		"name", h.Name, "coefficient", h.Coefficient, "type", h.Type,
	)
	return h, nil
}

// Update cập nhật ngày lễ. Nếu date thay đổi → kiểm tra uniqueness lại.
func (u *holidayUsecase) Update(ctx context.Context, req usecase.UpdateHolidayRequest) (*entity.Holiday, error) {
	existing, err := u.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	date, err := utils.ParseDateHCM(req.Date)
	if err != nil {
		return nil, apperrors.ErrHolidayInvalidDate
	}

	hType := req.Type
	if hType == "" {
		hType = existing.Type
	}
	if hType != entity.HolidayTypeNational && hType != entity.HolidayTypeCompany {
		return nil, apperrors.ErrHolidayInvalidType
	}

	coef := req.Coefficient
	if coef == 0 {
		coef = existing.Coefficient
	}
	if coef <= 0 || coef > 10 {
		return nil, apperrors.ErrHolidayInvalidCoefficient
	}

	if req.Name == "" {
		return nil, apperrors.NewValidationError(map[string]string{"name": "Tên ngày lễ không được để trống"})
	}

	var compensateFor *time.Time
	if req.IsCompensated {
		if req.CompensateFor == "" {
			return nil, apperrors.ErrHolidayCompensateMissing
		}
		cf, err := utils.ParseDateHCM(req.CompensateFor)
		if err != nil {
			return nil, apperrors.ErrHolidayCompensateMissing
		}
		compensateFor = &cf
	}

	if !sameDate(existing.Date, date) {
		exists, err := u.repo.ExistsByDate(ctx, date, existing.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperrors.ErrHolidayAlreadyExists
		}
	}

	// Cảnh báo nếu sửa ngày lễ đã qua — có thể ảnh hưởng summary đã xuất
	if existing.Date.Before(utils.Today()) {
		slog.Warn("holiday updated on past date — downstream summary/payroll may need re-check",
			"id", existing.ID, "date", existing.Date.Format("2006-01-02"),
		)
	}

	existing.Name = req.Name
	existing.Date = date
	existing.Year = date.Year()
	existing.Coefficient = coef
	existing.Type = hType
	existing.IsCompensated = req.IsCompensated
	existing.CompensateFor = compensateFor
	existing.Description = req.Description

	if err := u.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	u.invalidateCache(ctx)
	slog.Info("holiday updated", "id", existing.ID, "date", existing.Date.Format("2006-01-02"))
	return existing, nil
}

// Delete soft-delete ngày lễ.
func (u *holidayUsecase) Delete(ctx context.Context, id uint) error {
	existing, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.Date.Before(utils.Today()) {
		slog.Warn("holiday deleted on past date — downstream summary/payroll may need re-check",
			"id", existing.ID, "date", existing.Date.Format("2006-01-02"),
		)
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	u.invalidateCache(ctx)
	slog.Info("holiday deleted", "id", id, "date", existing.Date.Format("2006-01-02"))
	return nil
}

func (u *holidayUsecase) GetByID(ctx context.Context, id uint) (*entity.Holiday, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *holidayUsecase) GetList(ctx context.Context, filter repository.HolidayFilter) ([]*entity.Holiday, int64, error) {
	return u.repo.FindAll(ctx, filter)
}

// GetCalendar lấy danh sách ngày lễ trong khoảng, cache theo key range.
func (u *holidayUsecase) GetCalendar(ctx context.Context, from, to time.Time) ([]*entity.Holiday, error) {
	if to.Before(from) {
		return nil, apperrors.NewValidationError(map[string]string{"date_to": "date_to phải sau date_from"})
	}

	key := cacheKeyRange + from.Format("2006-01-02") + ":" + to.Format("2006-01-02")
	var cached []*entity.Holiday
	if err := u.cache.Get(ctx, key, &cached); err == nil && cached != nil {
		return cached, nil
	}

	holidays, err := u.repo.FindByDateRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Cache — ignore error (cache failure should not break request)
	if cacheErr := u.cache.Set(ctx, key, holidays, cacheTTL); cacheErr != nil {
		slog.Debug("holiday cache set failed", "key", key, "error", cacheErr)
	}
	return holidays, nil
}

// GetByDate tra cứu 1 ngày (cache per-date).
func (u *holidayUsecase) GetByDate(ctx context.Context, date time.Time) (*entity.Holiday, error) {
	key := cacheKeyDate + date.Format("2006-01-02")
	var cached entity.Holiday
	if err := u.cache.Get(ctx, key, &cached); err == nil && cached.ID > 0 {
		return &cached, nil
	}

	h, err := u.repo.FindByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	if h != nil {
		if cacheErr := u.cache.Set(ctx, key, h, cacheTTL); cacheErr != nil {
			slog.Debug("holiday cache set failed", "key", key, "error", cacheErr)
		}
	}
	return h, nil
}

// invalidateCache xoá toàn bộ cache holiday — gọi khi Create/Update/Delete.
func (u *holidayUsecase) invalidateCache(ctx context.Context) {
	if err := u.cache.DeletePattern(ctx, cacheInvalidateAll); err != nil {
		slog.Warn("holiday cache invalidate failed", "pattern", cacheInvalidateAll, "error", err)
	}
}

// sameDate so sánh 2 time.Time có cùng ngày (bỏ qua giờ/TZ).
func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// FormatYear helper (dùng cho router param)
func FormatYear(y int) string {
	return strconv.Itoa(y)
}

// ErrMsgYearInvalid message validate year (exported để handler dùng)
var ErrMsgYearInvalid = fmt.Sprintf("Năm không hợp lệ")
