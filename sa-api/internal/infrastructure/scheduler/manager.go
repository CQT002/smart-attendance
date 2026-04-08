package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/utils"
)

// TaskFunc định nghĩa hàm thực thi của scheduler. Trả về số record xử lý và error.
type TaskFunc func(ctx context.Context) (int64, error)

// Manager quản lý các scheduler từ DB, kiểm tra mỗi phút và dispatch task khi đến giờ.
type Manager struct {
	schedulerRepo repository.SchedulerRepository
	tasks         map[string]TaskFunc
}

// NewManager tạo instance Manager
func NewManager(schedulerRepo repository.SchedulerRepository) *Manager {
	return &Manager{
		schedulerRepo: schedulerRepo,
		tasks:         make(map[string]TaskFunc),
	}
}

// Register đăng ký task function theo tên scheduler (khớp với schedulers.name trong DB)
func (m *Manager) Register(name string, fn TaskFunc) {
	m.tasks[name] = fn
}

// Start bắt đầu vòng lặp kiểm tra scheduler mỗi phút
func (m *Manager) Start() {
	go func() {
		slog.Info("scheduler manager started", "registered_tasks", len(m.tasks))

		// Kiểm tra ngay khi khởi động
		m.checkAndRun()

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			m.checkAndRun()
		}
	}()
}

// checkAndRun load tất cả scheduler active từ DB, kiểm tra cron match và chạy
func (m *Manager) checkAndRun() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	schedulers, err := m.schedulerRepo.FindAllActive(ctx)
	if err != nil {
		slog.Error("scheduler manager: failed to load schedulers", "error", err)
		return
	}

	now := utils.Now()
	for _, s := range schedulers {
		if !cronMatch(s.CronExpr, now) {
			continue
		}

		// Đã chạy trong phút này rồi → bỏ qua
		if s.LastRunAt != nil && s.LastRunAt.In(utils.HCM).Format("200601021504") == now.Format("200601021504") {
			continue
		}

		fn, ok := m.tasks[s.Name]
		if !ok {
			slog.Warn("scheduler manager: no task registered", "name", s.Name)
			continue
		}

		go m.runTask(s, fn)
	}
}

// runTask thực thi task và ghi trạng thái vào DB
func (m *Manager) runTask(s *entity.Scheduler, fn TaskFunc) {
	timeout := time.Duration(s.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	slog.Info("scheduler running", "name", s.Name)

	now := utils.Now()
	count, err := fn(ctx)

	s.LastRunAt = &now
	if err != nil {
		s.LastStatus = "failed"
		s.LastError = err.Error()
		slog.Error("scheduler task failed", "name", s.Name, "error", err)
	} else {
		s.LastStatus = "success"
		s.LastError = ""
		slog.Info("scheduler task completed", "name", s.Name, "processed", count)
	}

	// Ghi trạng thái — dùng context mới vì context cũ có thể đã timeout
	updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer updateCancel()
	if updateErr := m.schedulerRepo.Update(updateCtx, s); updateErr != nil {
		slog.Error("scheduler manager: failed to update status", "name", s.Name, "error", updateErr)
	}
}

// cronMatch kiểm tra thời điểm now có khớp với cron expression 5-field không.
// Format: minute hour day month weekday
// Hỗ trợ: số cụ thể, "*" (any), "*/N" (step)
func cronMatch(expr string, now time.Time) bool {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false
	}

	values := []int{
		now.Minute(),
		now.Hour(),
		now.Day(),
		int(now.Month()),
		int(now.Weekday()), // 0=Sunday
	}

	for i, field := range fields {
		if !fieldMatch(field, values[i]) {
			return false
		}
	}
	return true
}

// fieldMatch kiểm tra một field cron có khớp với value không
func fieldMatch(field string, value int) bool {
	if field == "*" {
		return true
	}

	// */N — step
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err != nil || step <= 0 {
			return false
		}
		return value%step == 0
	}

	// Danh sách: "1,15" hoặc range: "1-5"
	for _, part := range strings.Split(field, ",") {
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			lo, err1 := strconv.Atoi(bounds[0])
			hi, err2 := strconv.Atoi(bounds[1])
			if err1 != nil || err2 != nil {
				continue
			}
			if value >= lo && value <= hi {
				return true
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			if value == n {
				return true
			}
		}
	}

	return false
}

// FormatCronDescription trả về mô tả dễ đọc cho cron expression (dùng cho log/admin UI)
func FormatCronDescription(expr string) string {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return expr
	}
	return fmt.Sprintf("minute=%s hour=%s day=%s month=%s weekday=%s",
		fields[0], fields[1], fields[2], fields[3], fields[4])
}
