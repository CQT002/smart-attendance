package holiday

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/utils"
	"golang.org/x/sync/errgroup"
)

// DailyAttendance chi tiết công của 1 ngày (dùng trong summary response cho client).
type DailyAttendance struct {
	Date         string                   `json:"date"`          // YYYY-MM-DD
	Weekday      int                      `json:"weekday"`       // 0=CN, 1=T2, ... 6=T7
	Status       entity.AttendanceStatus  `json:"status"`        // status hiển thị cho ngày
	WorkHours    float64                  `json:"work_hours"`    // giờ làm thực tế
	CheckIn      *time.Time               `json:"check_in_time,omitempty"`
	CheckOut     *time.Time               `json:"check_out_time,omitempty"`
	IsHoliday    bool                     `json:"is_holiday"`
	HolidayID    *uint                    `json:"holiday_id,omitempty"`
	HolidayName  string                   `json:"holiday_name,omitempty"`
	HolidayType  entity.HolidayType       `json:"holiday_type,omitempty"`
	Coefficient  float64                  `json:"coefficient,omitempty"`    // hệ số lương nếu là holiday
	EffectiveMul float64                  `json:"effective_multiplier"`     // hệ số thực tế ngày này đóng góp vào lương
	IsCompensated bool                    `json:"is_compensated,omitempty"` // true nếu là ngày nghỉ bù
}

// AttendanceCalculationResult tổng hợp công cho 1 user trong khoảng thời gian.
type AttendanceCalculationResult struct {
	UserID                uint              `json:"user_id"`
	DateFrom              string            `json:"date_from"`
	DateTo                string            `json:"date_to"`
	TotalWorkDays         int               `json:"total_work_days"`          // regular + holiday work + paid holiday + leave (tính lương)
	RegularWorkDays       int               `json:"regular_work_days"`        // ngày công thường
	HolidayWorkDays       int               `json:"holiday_work_days"`        // ngày làm việc trong lễ (có check-in)
	PaidHolidayDays       int               `json:"paid_holiday_days"`        // ngày nghỉ lễ hưởng lương
	LeaveDays             int               `json:"leave_days"`               // ngày nghỉ phép approved
	AbsentDays            int               `json:"absent_days"`              // vắng mặt không phép
	TotalWorkHours        float64           `json:"total_work_hours"`
	TotalSalaryMultiplier float64           `json:"total_salary_multiplier"`  // tổng hệ số quy đổi cho kỳ lương
	Days                  []DailyAttendance `json:"days"`
}

// CalcDeps dependencies cho calculation service.
type CalcDeps struct {
	AttendanceRepo repository.AttendanceRepository
	LeaveRepo      repository.LeaveRepository
	ShiftRepo      repository.ShiftRepository
	HolidayRepo    repository.HolidayRepository
}

// Calculator service tính công theo lịch sử chấm công + ngày lễ.
type Calculator struct {
	deps CalcDeps
}

// NewCalculator tạo instance Calculator với DI.
func NewCalculator(deps CalcDeps) *Calculator {
	return &Calculator{deps: deps}
}

// CalculateAttendanceLog tính công 1 user trong khoảng [from, to] (date-only).
//
// Luồng:
//   1. Load attendance_logs (1 row/ngày) + leave_requests + holidays in range.
//   2. Duyệt từng ngày trong range → ra DailyAttendance tương ứng.
//   3. Tổng hợp TotalWorkDays, HolidayWorkDays, TotalSalaryMultiplier.
//
// Quy ước hệ số:
//   - Ngày thường có check-in/out đủ giờ: 1.0
//   - Ngày thường thiếu giờ: workHours / shiftHours (≤ 1)
//   - Ngày lễ làm việc: coefficient × (workHours / shiftHours)
//   - Ngày nghỉ lễ hưởng lương (không làm): 1.0
//   - Ngày nghỉ phép approved: 1.0 (được trả lương)
//   - Vắng không phép: 0
func (c *Calculator) CalculateAttendanceLog(
	ctx context.Context, userID uint, branchID uint, from, to time.Time,
) (*AttendanceCalculationResult, error) {
	from = utils.StartOfDay(from)
	to = utils.StartOfDay(to)

	// Load attendance_logs
	filter := repository.AttendanceFilter{
		UserID:   &userID,
		DateFrom: &from,
		DateTo:   &to,
		Page:     1,
		Limit:    1000,
	}
	logs, _, err := c.deps.AttendanceRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}
	logByDate := make(map[string]*entity.AttendanceLog, len(logs))
	for _, l := range logs {
		logByDate[l.Date.Format("2006-01-02")] = l
	}

	// Load leaves approved — để phân biệt paid_holiday vs absent khi không có log
	leaves, _, err := c.deps.LeaveRepo.FindAll(ctx, repository.LeaveFilter{
		UserID: &userID,
		Status: entity.LeaveStatusApproved,
		Page:   1,
		Limit:  1000,
	})
	if err != nil {
		return nil, err
	}
	leaveByDate := make(map[string]*entity.LeaveRequest, len(leaves))
	for _, l := range leaves {
		d := l.LeaveDate
		if !d.Before(from) && !d.After(to) {
			leaveByDate[d.Format("2006-01-02")] = l
		}
	}

	// Load holidays
	holidays, err := c.deps.HolidayRepo.FindByDateRange(ctx, from, to)
	if err != nil {
		return nil, err
	}
	holidayByDate := make(map[string]*entity.Holiday, len(holidays))
	for _, h := range holidays {
		holidayByDate[h.Date.Format("2006-01-02")] = h
	}

	// Load shift default của branch (để biết shift.WorkHours)
	shift, _ := c.deps.ShiftRepo.FindDefault(ctx, branchID)
	shiftHours := 8.0
	if shift != nil && shift.WorkHours > 0 {
		shiftHours = shift.WorkHours
	}

	// Build day-by-day
	result := &AttendanceCalculationResult{
		UserID:   userID,
		DateFrom: from.Format("2006-01-02"),
		DateTo:   to.Format("2006-01-02"),
	}

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		day := DailyAttendance{
			Date:    key,
			Weekday: int(d.Weekday()),
		}

		log := logByDate[key]
		leave := leaveByDate[key]
		h := holidayByDate[key]

		if h != nil {
			day.IsHoliday = true
			id := h.ID
			day.HolidayID = &id
			day.HolidayName = h.Name
			day.HolidayType = h.Type
			day.Coefficient = h.Coefficient
			day.IsCompensated = h.IsCompensated
		}

		switch {
		case log != nil && log.CheckInTime != nil:
			// Có check-in — có thể là ngày thường hoặc làm ngày lễ
			day.Status = log.Status
			day.WorkHours = log.WorkHours
			day.CheckIn = log.CheckInTime
			day.CheckOut = log.CheckOutTime

			fraction := fractionOfShift(log.WorkHours, shiftHours)
			if h != nil {
				// Làm việc ngày lễ
				day.Status = entity.StatusHolidayWork
				day.EffectiveMul = h.Coefficient * fraction
				result.HolidayWorkDays++
			} else {
				// Ngày thường
				day.EffectiveMul = fraction
				result.RegularWorkDays++
			}
			result.TotalWorkHours += log.WorkHours

		case leave != nil:
			// Có leave approved → nghỉ hưởng lương (dù có trùng holiday hay không)
			day.Status = entity.StatusLeave
			day.EffectiveMul = 1.0
			result.LeaveDays++

		case h != nil:
			// Holiday + không đi làm + không leave → paid holiday
			day.Status = entity.StatusPaidHoliday
			day.EffectiveMul = 1.0
			result.PaidHolidayDays++

		case d.Weekday() == time.Saturday || d.Weekday() == time.Sunday:
			// Cuối tuần — không tính công, không tính absent
			day.Status = ""
			day.EffectiveMul = 0

		default:
			// Ngày thường không có log, không có leave, không holiday → absent (nếu đã qua)
			if d.Before(utils.Today()) {
				day.Status = entity.StatusAbsent
				result.AbsentDays++
			}
			day.EffectiveMul = 0
		}

		result.TotalSalaryMultiplier += day.EffectiveMul
		result.Days = append(result.Days, day)
	}

	result.TotalWorkDays = result.RegularWorkDays + result.HolidayWorkDays +
		result.PaidHolidayDays + result.LeaveDays

	// Round multiplier 2 decimals
	result.TotalSalaryMultiplier = roundTo(result.TotalSalaryMultiplier, 2)
	result.TotalWorkHours = roundTo(result.TotalWorkHours, 2)
	return result, nil
}

// BatchCalculate tính công song song cho nhiều user — dùng errgroup với semaphore giới hạn.
//
// Sử dụng cho chốt lương cuối tháng / export báo cáo admin.
// maxConcurrent: số goroutine đồng thời (mặc định 10 nếu <= 0).
func (c *Calculator) BatchCalculate(
	ctx context.Context,
	items []BatchCalcItem,
	maxConcurrent int,
) (map[uint]*AttendanceCalculationResult, error) {
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}

	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, maxConcurrent)

	var mu sync.Mutex
	results := make(map[uint]*AttendanceCalculationResult, len(items))

	for _, it := range items {
		it := it
		g.Go(func() error {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
			defer func() { <-sem }()

			r, err := c.CalculateAttendanceLog(ctx, it.UserID, it.BranchID, it.From, it.To)
			if err != nil {
				slog.Error("batch calculate failed", "user_id", it.UserID, "error", err)
				return err
			}
			mu.Lock()
			results[it.UserID] = r
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return results, err
	}
	return results, nil
}

// BatchCalcItem 1 entry cho BatchCalculate.
type BatchCalcItem struct {
	UserID   uint
	BranchID uint
	From     time.Time
	To       time.Time
}

// fractionOfShift = workHours / shiftHours, kẹp trong [0, 1].
func fractionOfShift(workHours, shiftHours float64) float64 {
	if shiftHours <= 0 {
		return 0
	}
	f := workHours / shiftHours
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

func roundTo(v float64, decimals int) float64 {
	p := 1.0
	for i := 0; i < decimals; i++ {
		p *= 10
	}
	return float64(int(v*p+0.5)) / p
}
