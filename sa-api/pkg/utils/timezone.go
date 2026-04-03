package utils

import (
	"log/slog"
	"time"
)

// HCM là timezone Asia/Ho_Chi_Minh (UTC+7) dùng cho toàn bộ hệ thống
var HCM *time.Location

func init() {
	var err error
	HCM, err = time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		slog.Warn("failed to load Asia/Ho_Chi_Minh timezone, using UTC+7 fixed offset", "error", err)
		HCM = time.FixedZone("ICT", 7*60*60)
	}
}

// Now trả về thời gian hiện tại theo timezone HCM
func Now() time.Time {
	return time.Now().In(HCM)
}

// Today trả về đầu ngày hôm nay theo timezone HCM (00:00:00)
func Today() time.Time {
	now := Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, HCM)
}

// StartOfDay trả về 00:00:00 của ngày đã cho theo timezone HCM
func StartOfDay(t time.Time) time.Time {
	t = t.In(HCM)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, HCM)
}

// EndOfDay trả về 23:59:59 của ngày đã cho theo timezone HCM
func EndOfDay(t time.Time) time.Time {
	t = t.In(HCM)
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, HCM)
}

// ParseDateHCM parse chuỗi "YYYY-MM-DD" thành time.Time theo timezone HCM
func ParseDateHCM(dateStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", dateStr, HCM)
}

// DateInHCM tạo time.Time từ year/month/day theo timezone HCM
func DateInHCM(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, HCM)
}
