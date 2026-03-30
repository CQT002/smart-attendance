package logger

import (
	"log/slog"
	"os"
)

// Setup khởi tạo slog logger theo môi trường
func Setup(env string, debug bool) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true, // Thêm file:line vào log để dễ debug
	}

	if debug {
		opts.Level = slog.LevelDebug
	}

	if env == "production" {
		// Production: JSON format để dễ parse bởi log aggregation tools (ELK, Loki)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Development: Text format dễ đọc
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// WithRequestID tạo logger với request ID để trace request
func WithRequestID(requestID string) *slog.Logger {
	return slog.With("request_id", requestID)
}

// WithUserID tạo logger với user ID
func WithUserID(userID uint) *slog.Logger {
	return slog.With("user_id", userID)
}

// WithBranchID tạo logger với branch ID
func WithBranchID(branchID uint) *slog.Logger {
	return slog.With("branch_id", branchID)
}
