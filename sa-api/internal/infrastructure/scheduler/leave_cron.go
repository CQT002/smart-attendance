package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/utils"
)

// StartLeaveAutoReject khởi chạy goroutine tự động reject yêu cầu nghỉ phép PENDING của tháng cũ.
// Chạy kiểm tra mỗi giờ, chỉ thực thi vào ngày 1 hàng tháng lúc 00:00–00:59 (HCM timezone).
func StartLeaveAutoReject(leaveUC usecase.LeaveUsecase) {
	go func() {
		slog.Info("leave auto-reject scheduler started")

		now := utils.Now()
		nextFirst := time.Date(now.Year(), now.Month()+1, 1, 0, 5, 0, 0, utils.HCM)
		if now.Day() == 1 && now.Hour() == 0 && now.Minute() < 5 {
			nextFirst = now
		}

		waitDuration := time.Until(nextFirst)
		if waitDuration > 0 {
			slog.Info("leave auto-reject: waiting for next execution",
				"next_run", nextFirst.Format(time.RFC3339),
				"wait", waitDuration.Round(time.Minute),
			)
			time.Sleep(waitDuration)
		}

		runLeaveAutoReject(leaveUC)

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		lastRunMonth := utils.Now().Month()

		for range ticker.C {
			now := utils.Now()
			if now.Day() == 1 && now.Month() != lastRunMonth {
				runLeaveAutoReject(leaveUC)
				lastRunMonth = now.Month()
			}
		}
	}()
}

func runLeaveAutoReject(leaveUC usecase.LeaveUsecase) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count, err := leaveUC.AutoRejectExpired(ctx)
	if err != nil {
		slog.Error("leave auto-reject failed", "error", err)
		return
	}

	slog.Info("leave auto-reject completed", "rejected_count", count)
}
