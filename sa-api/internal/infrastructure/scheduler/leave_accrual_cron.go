package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/utils"
)

// StartLeaveAccrual khởi chạy goroutine tự động cộng 1 ngày phép cho mỗi user active.
// Chạy vào 00:30 ngày 1 hàng tháng (HCM timezone).
func StartLeaveAccrual(leaveUC usecase.LeaveUsecase) {
	go func() {
		slog.Info("leave accrual scheduler started")

		// Tính thời điểm 00:30 ngày 1 tháng tới
		now := utils.Now()
		nextFirst := time.Date(now.Year(), now.Month()+1, 1, 0, 30, 0, 0, utils.HCM)
		if now.Day() == 1 && now.Hour() == 0 && now.Minute() < 30 {
			// Đang trong window ngày 1 → chạy ngay khi đến 00:30
			nextFirst = time.Date(now.Year(), now.Month(), 1, 0, 30, 0, 0, utils.HCM)
		}

		// Chờ đến thời điểm đầu tiên
		waitDuration := time.Until(nextFirst)
		if waitDuration > 0 {
			slog.Info("leave accrual: waiting for next execution",
				"next_run", nextFirst.Format(time.RFC3339),
				"wait", waitDuration.Round(time.Minute),
			)
			time.Sleep(waitDuration)
		}

		// Chạy lần đầu
		runLeaveAccrual(leaveUC)

		// Sau đó kiểm tra mỗi giờ
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		lastRunMonth := utils.Now().Month()

		for range ticker.C {
			now := utils.Now()
			// Chỉ chạy vào ngày 1, và chỉ chạy 1 lần mỗi tháng
			if now.Day() == 1 && now.Month() != lastRunMonth {
				runLeaveAccrual(leaveUC)
				lastRunMonth = now.Month()
			}
		}
	}()
}

func runLeaveAccrual(leaveUC usecase.LeaveUsecase) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count, err := leaveUC.AccrueMonthlyLeave(ctx)
	if err != nil {
		slog.Error("leave accrual failed", "error", err)
		return
	}

	slog.Info("leave accrual completed", "users_updated", count)
}
