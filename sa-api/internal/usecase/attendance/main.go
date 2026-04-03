package attendance

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
	maxSuspiciousCount = 3               // Số lần vi phạm tối đa trong 7 ngày
	todayCacheTTL      = 5 * time.Minute // TTL cache trạng thái hôm nay
)

type attendanceUsecase struct {
	attendanceRepo repository.AttendanceRepository
	userRepo       repository.UserRepository // để lấy branchID chính thống từ profile
	wifiConfigRepo repository.WiFiConfigRepository
	gpsConfigRepo  repository.GPSConfigRepository
	shiftRepo      repository.ShiftRepository
	cache          cache.Cache
}

// NewAttendanceUsecase tạo instance AttendanceUsecase
func NewAttendanceUsecase(
	attendanceRepo repository.AttendanceRepository,
	userRepo repository.UserRepository,
	wifiConfigRepo repository.WiFiConfigRepository,
	gpsConfigRepo repository.GPSConfigRepository,
	shiftRepo repository.ShiftRepository,
	cache cache.Cache,
) usecase.AttendanceUsecase {
	return &attendanceUsecase{
		attendanceRepo: attendanceRepo,
		userRepo:       userRepo,
		wifiConfigRepo: wifiConfigRepo,
		gpsConfigRepo:  gpsConfigRepo,
		shiftRepo:      shiftRepo,
		cache:          cache,
	}
}

// CheckIn xử lý nghiệp vụ check-in
//
// Flow:
//  1. Anti-fraud (FakeGPS flag + VPN flag từ SDK + server-side IP blocklist)
//  2. Lấy branchID từ profile user trong DB — không tin req.BranchID từ client
//  3. Kiểm tra đã check-in hôm nay chưa
//  4. Xác thực vị trí: WiFi (ưu tiên) → GPS Geofencing (fallback)
//  5. Lấy ca làm việc mặc định → tính trạng thái đi muộn
//  6. Tạo bản ghi và SET cache ngay để GetMyToday kế tiếp hit cache
func (u *attendanceUsecase) CheckIn(ctx context.Context, req usecase.CheckInRequest) (*entity.AttendanceLog, error) {
	logger := slog.With("user_id", req.UserID, "device_id", req.DeviceID, "ip", req.IPAddress)

	// === Bước 1: Anti-fraud ===
	// Kiểm tra flag từ mobile SDK + server-side IP blocklist
	if err := u.antiFraudCheck(ctx, req.UserID, req.IsFakeGPS, req.IsVPN, req.IPAddress, req.DeviceID); err != nil {
		logger.Warn("anti-fraud check failed", "error", err)
		return nil, err
	}

	// === Bước 2: Lấy branchID từ profile user — không tin client ===
	// Lý do: nếu dùng req.BranchID, nhân viên có thể khai branch khác để qua geofencing
	user, err := u.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user.BranchID == nil {
		// Admin không gắn chi nhánh cụ thể — không áp dụng chấm công theo chi nhánh
		return nil, apperrors.New(403, "NO_BRANCH", "Tài khoản không được gắn chi nhánh để chấm công")
	}
	branchID := *user.BranchID
	logger = logger.With("branch_id", branchID)

	// === Bước 3: Tìm record hôm nay ===
	today := utils.Now()
	existing, err := u.attendanceRepo.FindByUserAndDate(ctx, req.UserID, today)
	if err != nil {
		return nil, err
	}

	// Đã check-in rồi → trả về record hiện tại (giữ lần check-in đầu tiên)
	if existing != nil && existing.IsCheckedIn() {
		logger.Info("already checked in today, returning existing record", "attendance_id", existing.ID)
		return existing, nil
	}

	// === Bước 4: Xác thực vị trí theo chi nhánh của user ===
	checkMethod, err := u.validateLocation(ctx, branchID, req.Latitude, req.Longitude, req.SSID, req.BSSID)
	if err != nil {
		logger.Warn("location validation failed",
			"ssid", req.SSID,
			"lat", req.Latitude,
			"lng", req.Longitude,
		)
		return nil, err
	}

	// === Bước 5: Lấy ca mặc định → tính trạng thái ===
	shift, err := u.shiftRepo.FindDefault(ctx, branchID)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		shift = &entity.Shift{
			StartTime:   "08:00",
			EndTime:     "17:00",
			LateAfter:   15,
			EarlyBefore: 15,
			WorkHours:   8,
		}
	}

	now := utils.Now()

	// === Bước 6: Tạo hoặc cập nhật bản ghi ===
	if existing != nil {
		// Record đã tồn tại (checkout trước, chưa check-in) → cập nhật check-in vào record hiện tại
		existing.CheckInTime = &now
		existing.CheckInLat = &req.Latitude
		existing.CheckInLng = &req.Longitude
		existing.CheckInMethod = checkMethod
		existing.CheckInSSID = req.SSID
		existing.CheckInBSSID = req.BSSID
		existing.DeviceID = req.DeviceID
		existing.DeviceModel = req.DeviceModel
		existing.IPAddress = req.IPAddress
		existing.AppVersion = req.AppVersion
		existing.IsFakeGPS = req.IsFakeGPS
		existing.IsVPN = req.IsVPN
		existing.Status = u.calculateCheckInStatus(now, shift)

		if shift.ID != 0 {
			existing.ShiftID = &shift.ID
		}

		// Tính lại work hours nếu đã có checkout
		if existing.CheckOutTime != nil {
			workHours := existing.CheckOutTime.Sub(now).Hours()
			if workHours < 0 {
				workHours = 0
			}
			existing.WorkHours = roundToTwoDecimal(workHours)
		}

		if err := u.attendanceRepo.Update(ctx, existing); err != nil {
			return nil, err
		}

		cacheKey := u.todayCacheKey(req.UserID)
		u.cache.Set(ctx, cacheKey, existing, todayCacheTTL)
		logger.Info("check-in updated on existing record", "attendance_id", existing.ID)
		return existing, nil
	}

	// Chưa có record → tạo mới
	attendLog := &entity.AttendanceLog{
		UserID:        req.UserID,
		BranchID:      branchID,
		Date:          today,
		CheckInTime:   &now,
		CheckInLat:    &req.Latitude,
		CheckInLng:    &req.Longitude,
		CheckInMethod: checkMethod,
		CheckInSSID:   req.SSID,
		CheckInBSSID:  req.BSSID,
		DeviceID:      req.DeviceID,
		DeviceModel:   req.DeviceModel,
		IPAddress:     req.IPAddress,
		AppVersion:    req.AppVersion,
		IsFakeGPS:     req.IsFakeGPS,
		IsVPN:         req.IsVPN,
		Status:        u.calculateCheckInStatus(now, shift),
	}

	if shift.ID != 0 {
		attendLog.ShiftID = &shift.ID
	}

	if req.IsFakeGPS || req.IsVPN {
		note := ""
		if req.IsFakeGPS {
			note += "GPS giả được phát hiện. "
		}
		if req.IsVPN {
			note += "VPN/Proxy được phát hiện."
		}
		attendLog.FraudNote = note
		logger.Warn("suspicious check-in recorded", "fraud_note", note)
	}

	if err := u.attendanceRepo.Create(ctx, attendLog); err != nil {
		return nil, err
	}

	// === Bước 7: SET cache ngay — GetMyToday kế tiếp hit cache, không cần query DB ===
	cacheKey := u.todayCacheKey(req.UserID)
	u.cache.Set(ctx, cacheKey, attendLog, todayCacheTTL)

	logger.Info("check-in successful", "attendance_id", attendLog.ID, "status", attendLog.Status, "method", checkMethod)
	return attendLog, nil
}

// CheckOut xử lý nghiệp vụ check-out và tính toán giờ làm.
// Cho phép check-out nhiều lần (luôn cập nhật lần cuối).
// Cho phép check-out mà chưa check-in (tạo record mới nếu cần).
func (u *attendanceUsecase) CheckOut(ctx context.Context, req usecase.CheckOutRequest) (*entity.AttendanceLog, error) {
	logger := slog.With("user_id", req.UserID)

	// === Bước 1: Anti-fraud ===
	if req.IsFakeGPS {
		logger.Warn("fake GPS detected on check-out - allowing with flag")
	}
	if req.IsVPN {
		logger.Warn("VPN detected on check-out - allowing with flag")
	}
	if err := u.checkIPBlocklist(ctx, req.IPAddress); err != nil {
		logger.Warn("checkout blocked - ip in blocklist", "ip", req.IPAddress)
		return nil, err
	}

	// === Bước 2: Lấy branchID từ profile user ===
	user, err := u.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user.BranchID == nil {
		return nil, apperrors.New(403, "NO_BRANCH", "Tài khoản không được gắn chi nhánh để chấm công")
	}
	branchID := *user.BranchID

	// === Bước 3: Xác thực vị trí ===
	checkMethod, err := u.validateLocation(ctx, branchID, req.Latitude, req.Longitude, req.SSID, req.BSSID)
	if err != nil {
		return nil, err
	}

	// === Bước 4: Tìm hoặc tạo record hôm nay ===
	today := utils.Now()
	attendLog, err := u.attendanceRepo.FindByUserAndDate(ctx, req.UserID, today)
	if err != nil {
		return nil, err
	}

	now := utils.Now()

	if attendLog == nil {
		// Chưa có record hôm nay (chưa check-in) → tạo record mới chỉ với check-out
		shift, _ := u.shiftRepo.FindDefault(ctx, branchID)

		attendLog = &entity.AttendanceLog{
			UserID:         req.UserID,
			BranchID:       branchID,
			Date:           today,
			CheckOutTime:   &now,
			CheckOutLat:    &req.Latitude,
			CheckOutLng:    &req.Longitude,
			CheckOutMethod: checkMethod,
			CheckOutSSID:   req.SSID,
			CheckOutBSSID:  req.BSSID,
			DeviceID:       req.DeviceID,
			IPAddress:      req.IPAddress,
			IsFakeGPS:      req.IsFakeGPS,
			IsVPN:          req.IsVPN,
			Status:         entity.StatusPresent,
		}
		if shift != nil {
			attendLog.ShiftID = &shift.ID
		}

		if err := u.attendanceRepo.Create(ctx, attendLog); err != nil {
			return nil, err
		}
	} else {
		// Đã có record → cập nhật check-out (luôn lấy lần cuối)
		attendLog.CheckOutTime = &now
		attendLog.CheckOutLat = &req.Latitude
		attendLog.CheckOutLng = &req.Longitude
		attendLog.CheckOutMethod = checkMethod
		attendLog.CheckOutSSID = req.SSID
		attendLog.CheckOutBSSID = req.BSSID

		if attendLog.CheckInTime != nil {
			workHours := now.Sub(*attendLog.CheckInTime).Hours()
			attendLog.WorkHours = roundToTwoDecimal(workHours)

			if attendLog.ShiftID != nil {
				shift, err := u.shiftRepo.FindByID(ctx, *attendLog.ShiftID)
				if err == nil && shift != nil {
					attendLog.Overtime = u.calculateOvertime(workHours, shift)
					attendLog.Status = u.calculateCheckOutStatus(attendLog.Status, now, shift)
				}
			}
		}

		if err := u.attendanceRepo.Update(ctx, attendLog); err != nil {
			return nil, err
		}
	}

	// === Bước 5: SET cache ===
	cacheKey := u.todayCacheKey(req.UserID)
	u.cache.Set(ctx, cacheKey, attendLog, todayCacheTTL)

	logger.Info("check-out successful",
		"attendance_id", attendLog.ID,
		"work_hours", attendLog.WorkHours,
		"method", checkMethod,
	)
	return attendLog, nil
}

func (u *attendanceUsecase) GetByID(ctx context.Context, id uint) (*entity.AttendanceLog, error) {
	return u.attendanceRepo.FindByID(ctx, id)
}

func (u *attendanceUsecase) GetList(ctx context.Context, filter repository.AttendanceFilter) ([]*entity.AttendanceLog, int64, error) {
	return u.attendanceRepo.FindAll(ctx, filter)
}

// GetMyToday lấy bản ghi chấm công hôm nay, ưu tiên Redis cache
func (u *attendanceUsecase) GetMyToday(ctx context.Context, userID uint) (*entity.AttendanceLog, error) {
	cacheKey := u.todayCacheKey(userID)

	var cached entity.AttendanceLog
	if err := u.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	// Cache miss — query DB và populate cache
	log, err := u.attendanceRepo.FindByUserAndDate(ctx, userID, utils.Now())
	if err != nil {
		return nil, err
	}
	if log != nil {
		u.cache.Set(ctx, cacheKey, log, todayCacheTTL)
	}

	return log, nil
}

func (u *attendanceUsecase) GetSummary(ctx context.Context, userID uint, from, to time.Time) (*repository.AttendanceSummary, error) {
	return u.attendanceRepo.GetSummary(ctx, userID, from, to)
}

// ===================== Private helpers =====================

// antiFraudCheck kiểm tra gian lận nhiều lớp:
//  1. Flag GPS giả từ mobile SDK
//  2. Flag VPN từ mobile SDK
//  3. Server-side: IP có trong danh sách VPN/Proxy bị chặn
//  4. Lịch sử vi phạm: đã bị flag >= 3 lần trong 7 ngày
func (u *attendanceUsecase) antiFraudCheck(ctx context.Context, userID uint, isFakeGPS, isVPN bool, ip, deviceID string) error {
	// Fake GPS và VPN: chỉ log warning, cho phép check-in nhưng đánh dấu để quản lý review
	if isFakeGPS {
		slog.Warn("fake GPS detected - allowing check-in with flag", "user_id", userID)
	}
	if isVPN {
		slog.Warn("VPN detected - allowing check-in with flag", "user_id", userID)
	}

	// Server-side IP verification — không phụ thuộc vào flag từ client
	if err := u.checkIPBlocklist(ctx, ip); err != nil {
		return err
	}

	// Kiểm tra lịch sử vi phạm 7 ngày gần đây
	sevenDaysAgo := utils.Now().AddDate(0, 0, -7)
	count, err := u.attendanceRepo.CountSuspicious(ctx, userID, sevenDaysAgo)
	if err != nil {
		slog.Error("failed to check suspicious count", "user_id", userID, "error", err)
		return nil // fail-open: không block user nếu query lỗi
	}
	if count >= maxSuspiciousCount {
		slog.Warn("user blocked - repeated suspicious activity",
			"user_id", userID,
			"suspicious_count", count,
		)
		return apperrors.ErrSuspiciousActivity
	}

	return nil
}

// checkIPBlocklist kiểm tra IP có trong danh sách VPN/Proxy bị chặn không.
// Danh sách được quản lý bởi admin qua Redis key: blocked:ip:{ip}
// Fail-open: nếu Redis lỗi thì không block để tránh ảnh hưởng production.
func (u *attendanceUsecase) checkIPBlocklist(ctx context.Context, ip string) error {
	if ip == "" {
		return nil
	}
	exists, err := u.cache.Exists(ctx, "blocked:ip:"+ip)
	if err != nil {
		slog.Error("ip blocklist check error", "ip", ip, "error", err)
		return nil // fail-open
	}
	if exists {
		slog.Warn("check-in blocked - ip in blocklist", "ip", ip)
		return apperrors.ErrVPNDetected
	}
	return nil
}

// validateLocation xác thực vị trí qua WiFi (ưu tiên) hoặc GPS Geofencing (fallback).
// branchID phải là chi nhánh gốc của user lấy từ DB — không dùng giá trị từ client.
func (u *attendanceUsecase) validateLocation(
	ctx context.Context,
	branchID uint,
	lat, lng float64,
	ssid, bssid string,
) (*entity.CheckMethod, error) {
	// --- WiFi: nhanh và chính xác trong tòa nhà ---
	if ssid != "" || bssid != "" {
		valid, err := u.wifiConfigRepo.ValidateWiFi(ctx, branchID, ssid, bssid)
		if err != nil {
			return nil, err
		}
		if valid {
			method := entity.CheckMethodWiFi
			return &method, nil
		}
	}

	// --- GPS Geofencing: fallback khi không có WiFi khớp ---
	if lat != 0 || lng != 0 {
		if !utils.IsValidCoordinate(lat, lng) {
			return nil, apperrors.ErrLocationNotAllowed
		}

		gpsConfigs, err := u.gpsConfigRepo.FindActiveBranch(ctx, branchID)
		if err != nil {
			return nil, err
		}
		for _, cfg := range gpsConfigs {
			if utils.IsWithinGeofence(lat, lng, cfg.Latitude, cfg.Longitude, cfg.Radius) {
				method := entity.CheckMethodGPS
				return &method, nil
			}
		}
	}

	return nil, apperrors.ErrLocationNotAllowed
}

// calculateCheckInStatus tính trạng thái dựa vào giờ vào so với ca làm việc
func (u *attendanceUsecase) calculateCheckInStatus(checkInTime time.Time, shift *entity.Shift) entity.AttendanceStatus {
	checkInTime = checkInTime.In(utils.HCM)
	startHour, startMin := parseTime(shift.StartTime)
	shiftStart := time.Date(
		checkInTime.Year(), checkInTime.Month(), checkInTime.Day(),
		startHour, startMin, 0, 0, utils.HCM,
	)
	if checkInTime.After(shiftStart.Add(time.Duration(shift.LateAfter) * time.Minute)) {
		return entity.StatusLate
	}
	return entity.StatusPresent
}

// calculateCheckOutStatus cập nhật trạng thái nếu về sớm hơn quy định
// Kết hợp: late + early_leave → late_early_leave
func (u *attendanceUsecase) calculateCheckOutStatus(current entity.AttendanceStatus, checkOutTime time.Time, shift *entity.Shift) entity.AttendanceStatus {
	checkOutTime = checkOutTime.In(utils.HCM)
	endHour, endMin := parseTime(shift.EndTime)
	shiftEnd := time.Date(
		checkOutTime.Year(), checkOutTime.Month(), checkOutTime.Day(),
		endHour, endMin, 0, 0, utils.HCM,
	)
	isEarlyLeave := checkOutTime.Before(shiftEnd.Add(-time.Duration(shift.EarlyBefore) * time.Minute))
	if isEarlyLeave {
		if current == entity.StatusLate {
			return entity.StatusLateEarlyLeave // Đi trễ + Về sớm
		}
		return entity.StatusEarlyLeave
	}
	return current
}

// calculateOvertime tính giờ làm thêm
func (u *attendanceUsecase) calculateOvertime(workHours float64, shift *entity.Shift) float64 {
	if workHours <= shift.WorkHours {
		return 0
	}
	return roundToTwoDecimal(workHours - shift.WorkHours)
}

// todayCacheKey trả về Redis key cho trạng thái chấm công hôm nay của user
func (u *attendanceUsecase) todayCacheKey(userID uint) string {
	return cache.BuildKey(cache.KeyPrefixAttend, fmt.Sprintf("today:%d", userID))
}

func parseTime(t string) (int, int) {
	if len(t) != 5 {
		return 8, 0
	}
	hour, _ := strconv.Atoi(t[:2])
	min, _ := strconv.Atoi(t[3:])
	return hour, min
}

func roundToTwoDecimal(f float64) float64 {
	return float64(int(f*100)) / 100
}
