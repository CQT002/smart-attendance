package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/hdbank/smart-attendance/internal/domain/usecase"
	"github.com/hdbank/smart-attendance/pkg/utils"
)

// StartCorrectionAutoReject khởi chạy goroutine tự động reject yêu cầu PENDING của tháng cũ.
// Chạy kiểm tra mỗi giờ, chỉ thực thi vào ngày 1 hàng tháng lúc 00:00–00:59 (HCM timezone).
// Dùng idempotent UPDATE nên chạy nhiều lần cũng không sao.
func StartCorrectionAutoReject(correctionUC usecase.CorrectionUsecase) {
	go func() {
		slog.Info("correction auto-reject scheduler started")

		// Tính thời điểm 00:05 ngày 1 tháng tới
		now := utils.Now()
		nextFirst := time.Date(now.Year(), now.Month()+1, 1, 0, 5, 0, 0, utils.HCM)
		if now.Day() == 1 && now.Hour() == 0 && now.Minute() < 5 {
			// Đang trong window ngày 1 → chạy ngay
			nextFirst = now
		}

		// Chờ đến thời điểm đầu tiên
		waitDuration := time.Until(nextFirst)
		if waitDuration > 0 {
			slog.Info("correction auto-reject: waiting for next execution",
				"next_run", nextFirst.Format(time.RFC3339),
				"wait", waitDuration.Round(time.Minute),
			)
			time.Sleep(waitDuration)
		}

		// Chạy lần đầu
		runAutoReject(correctionUC)

		// Sau đó kiểm tra mỗi giờ
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		lastRunMonth := utils.Now().Month()

		for range ticker.C {
			now := utils.Now()
			// Chỉ chạy vào ngày 1, và chỉ chạy 1 lần mỗi tháng
			if now.Day() == 1 && now.Month() != lastRunMonth {
				runAutoReject(correctionUC)
				lastRunMonth = now.Month()
			}
		}
	}()
}

func runAutoReject(correctionUC usecase.CorrectionUsecase) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count, err := correctionUC.AutoRejectExpired(ctx)
	if err != nil {
		slog.Error("correction auto-reject failed", "error", err)
		return
	}

	slog.Info("correction auto-reject completed", "rejected_count", count)
}
