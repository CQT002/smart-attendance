# HDBBank Smart Attendance System

> Hệ thống chấm công thông minh cho **100 chi nhánh — 5.000 nhân viên** của HDBBank.
> Backend: **Go 1.22 + Echo v4** · Frontend: **Next.js 15 + TypeScript** · Mobile: **Flutter 3.16+ (BLoC)**

---

## Mục lục

1. [Tổng quan kiến trúc](#1-tổng-quan-kiến-trúc)
2. [Cơ chế Check-in Anti-fraud](#2-cơ-chế-check-in-anti-fraud)
3. [Kiến trúc Multi-branch & Data Isolation](#3-kiến-trúc-multi-branch--data-isolation)
4. [Hệ thống Phân quyền RBAC](#4-hệ-thống-phân-quyền-rbac)
5. [Chiến lược Scaling — 5.000 nhân viên đồng thời](#5-chiến-lược-scaling--5000-nhân-viên-đồng-thời)
6. [Git Flow & CI/CD](#6-git-flow--cicd)
7. [Folder Structure](#7-folder-structure)
8. [Cài đặt & Chạy Project](#8-cài-đặt--chạy-project)

---

## 1. Tổng quan kiến trúc

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT LAYER                             │
│  ┌──────────────────┐          ┌──────────────────────────────┐ │
│  │  sa-web          │          │  sa-mb                       │ │
│  │  Next.js 15      │          │  Flutter (BLoC)       │ │
│  │  Admin Portal    │          │  Employee App                │ │
│  └────────┬─────────┘          └───────────────┬──────────────┘ │
└───────────┼────────────────────────────────────┼───────────────┘
            │ HTTPS / REST JSON                  │ HTTPS / REST JSON
            ▼                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                       sa-api  (Go 1.22)                         │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │   Handler    │→ │   Usecase    │→ │    Repository      │    │
│  │  (Echo v4)   │  │ (Biz Logic)  │  │  (GORM/PostgreSQL) │    │
│  └──────────────┘  └──────────────┘  └────────────────────┘    │
│                                                                 │
│  ┌──────────────────────────┐  ┌───────────────────────────┐   │
│  │  Middleware Stack        │  │  Infrastructure            │   │
│  │  JWT · RBAC · RateLimit  │  │  Redis Cache · DB Pool     │   │
│  │  RequestLogger · CORS    │  │  slog · Viper Config       │   │
│  └──────────────────────────┘  └───────────────────────────┘   │
└─────────────────────────┬───────────────────────────┬──────────┘
                          │                           │
              ┌───────────▼──────────┐   ┌────────────▼──────────┐
              │  PostgreSQL 16       │   │  Redis 7              │
              │  Primary Database    │   │  Cache + Rate Limit   │
              │  B-tree Indexes      │   │  Session Store        │
              └──────────────────────┘   └───────────────────────┘
```

### Tech Stack tóm tắt

| Layer | Công nghệ | Lý do chọn |
|---|---|---|
| Backend | Go 1.22 + Echo v4 | Low-latency, goroutine-native concurrency |
| ORM | GORM v2 + `gorm.io/driver/postgres` | Type-safe, migration tích hợp |
| Database | PostgreSQL 16 | ACID, partial index, CTE, UNIQUE constraint |
| Cache | Redis 7 (`go-redis/v9`) | Atomic INCR cho rate limit, sub-ms latency |
| Auth | JWT HS256 (`golang-jwt/jwt/v5`) | Stateless, scalable ngang |
| Password | bcrypt cost 10 | Đủ an toàn, không quá chậm |
| Config | Viper + `config.yaml` | Hỗ trợ override qua env var |
| Logging | `log/slog` (stdlib) | Structured JSON log, zero external dep |
| Frontend | Next.js 15 + App Router | SSR-ready, route-group layout |
| UI | Tailwind CSS + shadcn/ui | Radix UI primitives, accessible |
| Data Fetching | TanStack Query v5 | Server-state caching, stale-while-revalidate |
| Forms | React Hook Form + Zod | Type-safe validation, minimal re-renders |
| Charts | Recharts | Composable, SSR-compatible |
| Mobile | Flutter 3.16+ (Dart) | Cross-platform, BLoC state management |
| Mobile HTTP | Dio | Interceptors, JWT auto-attach |
| Mobile GPS | Geolocator | Geofencing, mock detection |
| Mobile WiFi | network_info_plus | SSID/BSSID scan |
| Migration | gormigrate v2 | Versioned DB migration, auto-run on startup |

---

## 2. Cơ chế Check-in Anti-fraud

Đây là phần **phức tạp nhất** của hệ thống. Mỗi lần check-in phải vượt qua **3 lớp xác thực** tuần tự. Nếu bất kỳ lớp nào thất bại, request bị từ chối ngay lập tức.

### Sơ đồ luồng xử lý

```
POST /api/v1/attendance/check-in
            │
            ▼
┌───────────────────────┐
│  LAYER 0: Rate Limit  │── FAIL ──→ 429 Too Many Requests
│  10 req/min per user  │           (Redis sliding window)
└───────────┬───────────┘
            │ PASS
            ▼
┌───────────────────────────────────────────────────────┐
│  LAYER 1: Device Integrity (Anti-fraud)               │
│                                                       │
│  1a. is_fake_gps = true?  → ErrFakeGPSDetected        │
│  1b. is_vpn = true?       → ErrVPNDetected            │
│  1c. IP in Redis blocklist → ErrIPBlocked             │
│  1d. ≥3 violations/7 ngày → ErrSuspiciousActivity     │
└───────────────────────┬───────────────────────────────┘
            │ PASS
            ▼
┌───────────────────────────────────────────────────────┐
│  LAYER 2: Location Validation                         │
│                                                       │
│  Priority: WiFi trước, GPS fallback                   │
│                                                       │
│  2a. WiFi (BSSID/SSID):                              │
│      SELECT FROM wifi_configs                         │
│      WHERE branch_id = ? AND (ssid=? OR bssid=?)     │
│      → MATCH → check_in_method = 'wifi'              │
│                                                       │
│  2b. GPS Geofencing (Haversine):                      │
│      SELECT FROM gps_configs                          │
│      WHERE branch_id = ? AND is_active = true        │
│      FOR EACH config:                                 │
│        if distance(user, center) ≤ radius → PASS     │
│                                                       │
│  → No match → ErrLocationNotAllowed                   │
└───────────────────────┬───────────────────────────────┘
            │ PASS
            ▼
┌───────────────────────────────────────────────────────┐
│  LAYER 3: Business Rules                              │
│                                                       │
│  3a. Đã check-in hôm nay? (UNIQUE idx lookup)        │
│  3b. Tính trạng thái: present / late                 │
│      if now > shift.start + late_after → StatusLate  │
│  3c. Ghi DB + Set Redis cache (TTL 5 phút)           │
└───────────────────────────────────────────────────────┘
            │
            ▼
         201 Created + AttendanceLog
```

### Chi tiết từng lớp

#### Lớp 1 — Device Integrity

```go
// usecase/attendance/main.go
func (u *attendanceUsecase) antiFraudCheck(ctx, userID, isFakeGPS, isVPN, ip, deviceID) error {

    // 1a. Flag từ Mobile SDK (Android/iOS phát hiện mock location)
    if isFakeGPS {
        return apperrors.New(403, "FAKE_GPS_DETECTED", "Phát hiện GPS giả")
    }

    // 1b. Flag từ Mobile SDK (phát hiện VPN/Proxy active)
    if isVPN {
        return apperrors.New(403, "VPN_DETECTED", "Không được dùng VPN khi chấm công")
    }

    // 1c. Server-side IP blocklist (Admin có thể block IP qua Redis)
    if blocked, _ := u.cache.Exists(ctx, "blocked:ip:"+ip); blocked {
        return apperrors.New(403, "IP_BLOCKED", "IP bị chặn")
    }

    // 1d. Kiểm tra lịch sử vi phạm (partial index query)
    since := time.Now().AddDate(0, 0, -7)
    count, _ := u.attendanceRepo.CountSuspicious(ctx, userID, since)
    if count >= 3 {
        return apperrors.New(403, "SUSPICIOUS_ACTIVITY", "Tài khoản bị tạm khóa do vi phạm nhiều lần")
    }

    return nil
}
```

> **Thiết kế quan trọng:** Dù gian lận bị phát hiện ở bước 1d, hệ thống vẫn **cho phép check-in nhưng đánh dấu `is_fake_gps=true`** vào DB để manager xem xét — tránh block nhân viên oan trong edge case. Chỉ block khi đạt ngưỡng vi phạm.

#### Lớp 2 — Location Validation

**WiFi được ưu tiên** vì chính xác hơn trong nhà, không bị che khuất tín hiệu GPS:

```go
func (u *attendanceUsecase) validateLocation(ctx, branchID, lat, lng, ssid, bssid) (CheckMethod, error) {

    // Thử WiFi trước
    if ssid != "" || bssid != "" {
        valid, _ := u.wifiConfigRepo.ValidateWiFi(ctx, branchID, ssid, bssid)
        if valid {
            return CheckMethodWiFi, nil  // Xác thực thành công
        }
    }

    // Fallback sang GPS Geofencing
    if lat != 0 && lng != 0 {
        if !pkg.IsValidCoordinate(lat, lng) {
            return "", ErrLocationNotAllowed
        }
        configs, _ := u.gpsConfigRepo.FindActiveBranch(ctx, branchID)
        for _, cfg := range configs {
            if pkg.IsWithinGeofence(lat, lng, cfg.Latitude, cfg.Longitude, cfg.Radius) {
                return CheckMethodGPS, nil
            }
        }
    }

    return "", ErrLocationNotAllowed
}
```

**Công thức Haversine** tính khoảng cách trên bề mặt Trái Đất (độ chính xác ±10m):

```go
// pkg/utils/geo.go
func IsWithinGeofence(userLat, userLng, centerLat, centerLng, radiusM float64) bool {
    const R = 6371000 // bán kính Trái Đất (mét)
    φ1, φ2 := userLat*math.Pi/180, centerLat*math.Pi/180
    Δφ := (centerLat - userLat) * math.Pi / 180
    Δλ := (centerLng - userLng) * math.Pi / 180

    a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
         math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

    return R*c <= radiusM
}
```

#### Lớp 3 — Business Rules & Security

```go
// CRITICAL: Đọc branch_id từ DB, KHÔNG tin dữ liệu client
user, _ := u.userRepo.FindByID(ctx, req.UserID)
branchID := *user.BranchID  // ← Sử dụng giá trị từ DB

// Lý do: Prevent privilege escalation — nếu user gửi branch_id giả,
// họ có thể pass geofencing của chi nhánh khác.
// Đây là security decision quan trọng nhất trong luồng check-in.
```

### So sánh phương thức xác thực

| Tiêu chí | WiFi (BSSID/SSID) | GPS Geofencing |
|---|---|---|
| **Độ chính xác** | Cao (phòng cụ thể) | Trung bình (±10m) |
| **Hoạt động trong nhà** | Tốt | Kém (tín hiệu yếu) |
| **Có thể bị fake** | Khó (cần access point thật) | Dễ hơn (mock location) |
| **Phụ thuộc phần cứng** | WiFi enabled | GPS enabled |
| **Ưu tiên** | **1 (Primary)** | 2 (Fallback) |

---

## 3. Kiến trúc Multi-branch & Data Isolation

### Database Schema — Thiết kế cho 100 chi nhánh

```sql
-- Branches: Hub trung tâm, mỗi chi nhánh độc lập
CREATE TABLE branches (
    id      SERIAL PRIMARY KEY,
    code    VARCHAR(20) NOT NULL UNIQUE,  -- CN001, CN002...
    name    VARCHAR(200) NOT NULL,
    city    VARCHAR(100),
    ...
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX idx_branches_is_active ON branches (is_active);

-- Users: Gắn chặt với branch thông qua FK
CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    branch_id   INTEGER REFERENCES branches(id) ON DELETE SET NULL,
    --                 ↑ NULL chỉ cho Admin (không thuộc chi nhánh nào)
    role        VARCHAR(20) CHECK (role IN ('admin', 'manager', 'employee')),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    ...
    CONSTRAINT uq_user_email        UNIQUE (email),
    CONSTRAINT uq_user_employee_code UNIQUE (employee_code)
);

-- Index composite cho pattern query phổ biến nhất của Manager
CREATE INDEX idx_users_branch_role_active ON users (branch_id, role, is_active);
--                                                   ↑         ↑     ↑
--                                          Lọc theo CN  + Role + Active

-- Attendance: Mỗi bản ghi gắn với CÙNG LÚC user VÀ branch
CREATE TABLE attendance_logs (
    id        SERIAL PRIMARY KEY,
    user_id   INTEGER NOT NULL REFERENCES users(id),
    branch_id INTEGER NOT NULL REFERENCES branches(id),
    date      DATE    NOT NULL,
    ...
    CONSTRAINT uq_attendance_user_date UNIQUE (user_id, date)
    --         ↑ Database-level prevention của duplicate check-in
);

-- Index cho query check-in hôm nay (hot path, chạy mọi request)
CREATE INDEX idx_attendance_user_date    ON attendance_logs (user_id, date DESC);
CREATE INDEX idx_attendance_branch_date  ON attendance_logs (branch_id, date DESC);

-- Partial index chỉ cho records gian lận (giảm kích thước index đáng kể)
CREATE INDEX idx_attendance_fraud ON attendance_logs (user_id, created_at DESC)
    WHERE is_fake_gps = TRUE OR is_vpn = TRUE;
```

### DailySummary — Pre-computed Aggregates

Thay vì chạy `GROUP BY` tốn kém mỗi khi load dashboard, hệ thống duy trì bảng tổng hợp được tính trước:

```sql
CREATE TABLE daily_summaries (
    id              SERIAL PRIMARY KEY,
    branch_id       INTEGER NOT NULL REFERENCES branches(id),
    date            DATE    NOT NULL,
    total_employees INTEGER NOT NULL DEFAULT 0,
    present_count   INTEGER NOT NULL DEFAULT 0,
    late_count      INTEGER NOT NULL DEFAULT 0,
    absent_count    INTEGER NOT NULL DEFAULT 0,
    attendance_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    on_time_rate    DECIMAL(5,2) NOT NULL DEFAULT 0,
    ...
    CONSTRAINT uq_daily_branch_date UNIQUE (branch_id, date)
    -- Cho phép UPSERT: INSERT ... ON CONFLICT (branch_id, date) DO UPDATE SET ...
);

-- Range scan hiệu quả cho báo cáo 30 ngày
CREATE INDEX idx_daily_branch_date ON daily_summaries (branch_id, date DESC);
-- Full-system view (Admin): tất cả chi nhánh trong 1 ngày
CREATE INDEX idx_daily_date         ON daily_summaries (date);
```

### Data Isolation — Cách Manager chỉ thấy dữ liệu chi nhánh mình

Nguyên tắc: **branch isolation được thực thi ở 3 tầng đồng thời** — không phụ thuộc vào một điểm duy nhất:

```
Tầng 1 (Middleware)     Tầng 2 (Handler)         Tầng 3 (Repository)
─────────────────────   ──────────────────────   ─────────────────────────
JWTAuth đọc token       Handler kiểm tra         Query tự động thêm
→ ghi branch_id         req.BranchID vs          WHERE branch_id = ?
  vào Echo Context       context.branch_id        dựa trên filter
                         → ErrForbidden nếu       được truyền xuống
                           không khớp
```

**Ví dụ cụ thể** — Manager chi nhánh 5 gọi API lấy danh sách nhân viên:

```go
// handler/admin/user_handler.go
func (h *UserHandler) GetList(c echo.Context) error {
    role := middleware.GetRole(c)
    userBranchID := middleware.GetBranchID(c) // Lấy từ JWT token

    var filter repository.UserFilter
    // Parse query params: ?branch_id=1&role=employee...
    bindFilter(c, &filter)

    // Nếu là Manager: bắt buộc filter theo branch của họ,
    // bất kể client gửi branch_id nào trong query string
    if role == entity.RoleManager {
        filter.BranchID = userBranchID // Override bằng giá trị từ JWT
    }
    // Admin: giữ nguyên filter từ query string (full access)

    users, total, err := h.userUC.GetList(ctx, filter)
    ...
}

// repository/user_repo.go
func (r *userRepo) FindAll(ctx, filter UserFilter) ([]*entity.User, int64, error) {
    query := r.db.WithContext(ctx).Model(&entity.User{})

    if filter.BranchID != nil {
        query = query.Where("branch_id = ?", *filter.BranchID)
        // ↑ WHERE branch_id = 5 — Tự động giới hạn phạm vi
    }
    // ...
}
```

### CTE Query cho Dashboard — Không N+1

Report dashboard tổng hợp dùng CTE thay vì loop từng chi nhánh:

```sql
-- Một query duy nhất thay vì 100 queries
WITH employee_counts AS (
    SELECT branch_id, COUNT(*) AS total_employees
    FROM users
    WHERE is_active = true
    GROUP BY branch_id
),
today_stats AS (
    SELECT
        branch_id,
        COUNT(*) FILTER (WHERE status = 'present')         AS present_count,
        COUNT(*) FILTER (WHERE status = 'late')            AS late_count,
        COUNT(*) FILTER (WHERE status = 'absent')          AS absent_count,
        COUNT(*) FILTER (WHERE is_fake_gps OR is_vpn)      AS fraud_count,
        ROUND(
            COUNT(*) FILTER (WHERE status IN ('present','late')) * 100.0
            / NULLIF(COUNT(*), 0), 2
        )                                                  AS attendance_rate
    FROM attendance_logs
    WHERE date = CURRENT_DATE
    GROUP BY branch_id
)
SELECT
    b.id, b.name, b.code,
    COALESCE(ec.total_employees, 0)  AS total_employees,
    COALESCE(ts.present_count, 0)    AS present_count,
    COALESCE(ts.late_count, 0)       AS late_count,
    COALESCE(ts.absent_count, 0)     AS absent_count,
    COALESCE(ts.attendance_rate, 0)  AS attendance_rate,
    COALESCE(ts.fraud_count, 0)      AS fraud_count
FROM branches b
LEFT JOIN employee_counts ec ON ec.branch_id = b.id
LEFT JOIN today_stats ts      ON ts.branch_id = b.id
WHERE b.is_active = true
ORDER BY b.name
LIMIT $1 OFFSET $2;
```

---

## 4. Hệ thống Phân quyền RBAC

### Ma trận quyền

| Hành động | Admin | Manager | Employee |
|---|:---:|:---:|:---:|
| Xem tất cả chi nhánh | ✅ | ❌ (chỉ chi nhánh mình) | ❌ |
| Tạo / Xoá chi nhánh | ✅ | ❌ | ❌ |
| Tạo nhân viên | ✅ | ✅ (branch mình) | ❌ |
| Xoá / Vô hiệu nhân viên | ✅ | ❌ | ❌ |
| Reset mật khẩu nhân viên | ✅ | ❌ | ❌ |
| Xem dữ liệu chấm công | ✅ (tất cả) | ✅ (branch mình) | ✅ (chỉ bản thân) |
| Xem Dashboard tổng hệ thống | ✅ | ❌ | ❌ |
| Xem Dashboard chi nhánh | ✅ | ✅ (branch mình) | ❌ |
| Báo cáo tất cả chi nhánh | ✅ | ❌ | ❌ |
| Check-in / Check-out | ❌ | ❌ | ✅ |

### JWT Claims — Nguồn sự thật

```go
// pkg/utils/jwt.go
type Claims struct {
    UserID   uint             `json:"user_id"`
    BranchID *uint            `json:"branch_id"` // nil nếu là Admin
    Role     entity.UserRole  `json:"role"`
    jwt.RegisteredClaims                          // exp, iat, iss...
}
```

Token payload **không thể bị giả mạo** (signed HS256). Handler đọc `BranchID` từ token, không từ request body/query.

### Middleware RBAC — Thực thi tại Router Layer

```go
// internal/server/router.go
admin := e.Group("/api/v1/admin", middleware.JWTAuth(jwtSecret))

// Nhóm chỉ Admin + Manager thấy
protected := admin.Group("", middleware.RequireRole(entity.RoleAdmin, entity.RoleManager))
protected.GET("/users", userHandler.GetList)
protected.POST("/users", userHandler.Create)

// Nhóm chỉ Admin được thực hiện
adminOnly := admin.Group("", middleware.RequireRole(entity.RoleAdmin))
adminOnly.DELETE("/users/:id", userHandler.Delete)
adminOnly.POST("/users/:id/reset-password", userHandler.ResetPassword)
adminOnly.POST("/branches", branchHandler.Create)
adminOnly.DELETE("/branches/:id", branchHandler.Delete)
adminOnly.GET("/reports/branches", reportHandler.GetBranchReport)
```

```go
// middleware/auth.go
func RequireRole(allowedRoles ...entity.UserRole) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            role := GetRole(c)
            for _, allowed := range allowedRoles {
                if role == allowed {
                    return next(c) // PASS
                }
            }
            slog.Warn("access denied", "role", role, "path", c.Path())
            return apperrors.ErrForbidden
        }
    }
}
```

### RBAC trên Frontend (Next.js)

Route protection được thực hiện tại `app/(admin)/layout.tsx`:

```typescript
// app/(admin)/layout.tsx
export default function AdminLayout({ children }) {
    const router = useRouter();

    useEffect(() => {
        if (!isAuthenticated()) {
            router.replace("/login"); // Redirect nếu chưa đăng nhập
        }
    }, []);
    // ...
}
```

UI elements được ẩn/hiện theo role:

```typescript
// Ví dụ: Nút "Xoá chi nhánh" chỉ hiện với Admin
const { data: user } = useCurrentUser();

{user?.role === "admin" && (
    <Button variant="destructive" onClick={() => deleteBranch(id)}>
        Xoá
    </Button>
)}
```

---

## 5. Chiến lược Scaling — 5.000 nhân viên đồng thời

### Bài toán Peak Load

> **Kịch bản khắc nghiệt nhất:** 5.000 nhân viên check-in đồng loạt lúc 8:00 SA.
> Mỗi check-in kích hoạt: 1 Rate Limit check + 1 Fraud check + 1 WiFi/GPS query + 1 INSERT.

**Phân tích:**
- ~500 requests/phút trong 10 phút đầu giờ = ~8.3 req/giây
- Mỗi request cần ~3–5 DB queries + ~2 Redis ops

### Redis — Caching & Rate Limiting

#### Connection Pool

```go
// infrastructure/cache/redis.go
redis.NewClient(&redis.Options{
    Addr:         "redis:6379",
    PoolSize:     10,            // 10 concurrent connections
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
})
```

#### Key Strategy — Tránh Collision

```
Key Pattern                              TTL         Mục đích
────────────────────────────────────────────────────────────────
attend:today:{user_id}                  5 phút      Cache check-in hôm nay
dashboard:admin:stats                   5 phút      KPI tổng hệ thống
dashboard:today:branch:{id}:p{p}:l{l}  2 phút      Stats từng chi nhánh
rate:checkin:{user_id}                  1 phút      Counter rate limit check-in
rate:login:{ip}                         15 phút     Counter rate limit login
rate:global:{ip}                        1 phút      Global rate limit
blocked:ip:{ip}                         Dynamic     Blacklist IP
```

#### Rate Limiting — Sliding Window với Redis Atomic INCR

```go
// middleware/rate_limiter.go
func (m *RateLimiterMiddleware) checkLimit(ctx, key string, limit int, window time.Duration) error {
    count, err := m.cache.Incr(ctx, key) // Atomic increment
    if err != nil {
        slog.Error("rate limiter redis error", "error", err)
        return nil // Fail-open: nếu Redis lỗi, cho phép request
    }

    if count == 1 {
        m.cache.Expire(ctx, key, window) // Set TTL lần đầu
    }

    if count > int64(limit) {
        return apperrors.ErrRateLimitExceeded // 429
    }
    return nil
}
```

| Endpoint | Limit | Window | Key |
|---|---|---|---|
| `POST /attendance/check-in` | 10 req | 1 phút | `rate:checkin:{user_id}` |
| `POST /admin/auth/login` | 10 req | 15 phút | `rate:login:{ip}` |
| Tất cả routes | 100 req | 1 phút | `rate:global:{ip}` |

### PostgreSQL — Connection Pooling & Index Strategy

#### Connection Pool

```go
// infrastructure/database/postgres.go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)               // Tối đa 25 kết nối đồng thời
sqlDB.SetMaxIdleConns(10)               // Giữ 10 kết nối sẵn sàng trong pool
sqlDB.SetConnMaxLifetime(300 * time.Second) // Đóng kết nối sau 5 phút
```

| Tham số | Giá trị | Lý do |
|---|---|---|
| `MaxOpenConns` | 25 | Đủ cho 8 req/s, tránh overwhelm PostgreSQL |
| `MaxIdleConns` | 10 | Reuse connections, tránh overhead `CREATE CONNECTION` (~100ms) |
| `ConnMaxLifetime` | 300s | Refresh định kỳ, tránh stale connections |

#### Index Strategy — Từng Query Pattern

```
Truy vấn                         Index được dùng                    Chi phí
──────────────────────────────────────────────────────────────────────────────
FindByEmail (Login)              UNIQUE (email)                     O(log n)
FindByUserAndDate (Check-in)     UNIQUE (user_id, date)             O(log n)
GetAllActiveByBranch (Manager)   (branch_id, is_active)             O(log n + k)
GetFraudHistory (Anti-fraud)     PARTIAL (user_id) WHERE fraud=true O(log m), m<<n
ValidateWiFi (Check-in)          (branch_id, is_active) + bssid     O(log n)
GetShiftDefault (Check-in)       (branch_id, is_default, is_active) O(log n)
DashboardStats (Today)           (branch_id, date DESC)             O(log n + k)
```

**Partial Index** cho fraud detection — chỉ index các bản ghi vi phạm (thường < 1% tổng):

```sql
CREATE INDEX idx_attendance_fraud
    ON attendance_logs (user_id, created_at DESC)
    WHERE is_fake_gps = TRUE OR is_vpn = TRUE;
-- Size index: ~1% so với full index → query siêu nhanh
```

### Ước tính tải hệ thống

```
5.000 users × 2 actions/ngày (check-in + check-out) = 10.000 DB inserts/ngày
Giờ peak (8:00-8:30): ~500 inserts/30 phút = ~17 inserts/phút = 0.3 TPS

PostgreSQL 16 capacity: 1.000-10.000 TPS (simple inserts)
→ Hệ thống có headroom ~30-100x so với tải hiện tại

Redis capacity: ~100.000 ops/giây
→ Không bao giờ là bottleneck với 17 req/phút

Kết luận: Kiến trúc hiện tại scale được đến ~100.000 nhân viên
trước khi cần sharding hoặc read replicas.
```

### Chiến lược Scale Out (tương lai)

```
Giai đoạn 1 (hiện tại):   1 API server, 1 PostgreSQL, 1 Redis
Giai đoạn 2 (50k users):  N API servers (stateless JWT), PostgreSQL + Read Replica
                           Redis Sentinel (HA)
Giai đoạn 3 (500k users): PostgreSQL Sharding theo branch_id range
                           Redis Cluster, CDN cho static assets
```

---

## 6. Git Flow & CI/CD

### Branching Strategy

```
main ────────────────────────────────────────────── Production
  │
  └── develop ──────────────────────────────────── Staging (auto-deploy)
        │
        ├── feature/attendance-anti-fraud          Feature branches
        ├── feature/admin-dashboard
        ├── fix/jwt-token-refresh
        └── chore/upgrade-go-122
```

| Branch | Mục đích | Deployment |
|---|---|---|
| `main` | Production-ready code | Auto-deploy → Production |
| `develop` | Integration branch | Auto-deploy → Staging |
| `feature/*` | New features | PR → develop |
| `fix/*` | Bug fixes | PR → develop (hoặc main nếu hotfix) |
| `chore/*` | Deps, config, tooling | PR → develop |

### Conventional Commits

Mọi commit **bắt buộc** tuân theo format:

```
<type>(<scope>): <mô tả ngắn bằng tiếng Anh>
```

**Allowed types & scopes:**

```
Types:   feat | fix | refactor | perf | test | docs | chore | ci
Scopes:  auth | attendance | user | branch | report | middleware |
         config | db | cache | infra | handler | usecase | repository
```

**Ví dụ commit hợp lệ:**

```bash
feat(attendance): add WiFi + GPS dual validation on check-in
fix(auth): prevent timing attack on login with constant-time compare
perf(report): replace N+1 query with single CTE in GetTodayBranchStats
refactor(repository): extract base repository with common pagination
chore(config): migrate from .env to config.yaml with Viper
```

### Docker Compose (Development)

```yaml
# docker-compose.yml
version: "3.9"
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: smart_attendance
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports: ["5432:5432"]
    volumes: [postgres_data:/var/lib/postgresql/data]

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru

  api:
    build:
      context: ./sa-api
      dockerfile: Dockerfile
    ports: ["8080:8080"]
    environment:
      DATABASE_HOST: postgres
      REDIS_HOST: redis
      JWT_SECRET: ${JWT_SECRET}
    depends_on: [postgres, redis]

  web:
    build:
      context: ./sa-web
      dockerfile: Dockerfile
    ports: ["3000:3000"]
    environment:
      NEXT_PUBLIC_API_URL: http://api:8080/api/v1
    depends_on: [api]
```

### Dockerfile — Multi-stage Build (API)

```dockerfile
# sa-api/Dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# Stage 2: Runtime (minimal image ~15MB)
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config/config.yaml ./config/
EXPOSE 8080
CMD ["./server"]
```

---

## 7. Folder Structure

### Backend (`sa-api/`)

```
sa-api/
├── cmd/
│   ├── server/
│   │   └── main.go                 # Entry point: DI, auto-migration, graceful shutdown
│   └── migration/
│       └── main.go                 # CLI tool: -cmd up|down|rollback-to|reset
├── config/
│   ├── config.go                   # Viper loader, Config struct
│   └── config.yaml                 # Dev values (không commit secret thật)
├── internal/
│   ├── domain/                     # Layer trong cùng — ZERO external imports
│   │   ├── entity/                 # GORM models + business types
│   │   │   ├── user.go             # User, UserRole enum
│   │   │   ├── branch.go           # Branch + WiFiConfig + GPSConfig
│   │   │   ├── attendance.go       # AttendanceLog, AttendanceStatus enum
│   │   │   ├── shift.go            # Shift (giờ làm việc)
│   │   │   └── daily_summary.go    # Pre-computed daily aggregates
│   │   ├── repository/             # Repository interfaces (contracts)
│   │   │   ├── user_repo.go        # UserFilter, UserRepository interface
│   │   │   ├── branch_repo.go
│   │   │   ├── attendance_repo.go  # AttendanceFilter, AttendanceSummary
│   │   │   ├── wifi_config_repo.go
│   │   │   ├── gps_config_repo.go
│   │   │   └── shift_repo.go
│   │   └── usecase/                # Usecase interfaces + request/response DTOs
│   │       ├── user_uc.go          # LoginRequest, LoginResponse, CreateUserRequest
│   │       ├── attendance_uc.go    # CheckInRequest, TodayStatsFilter
│   │       └── report_uc.go        # DashboardStats, BranchAttendanceReport
│   ├── repository/                 # PostgreSQL implementations
│   │   ├── user_repo.go            # FindAll, FindByEmail, Create, Update...
│   │   ├── branch_repo.go
│   │   ├── attendance_repo.go      # GetTodayStatsByBranch (CTE query)
│   │   ├── wifi_config_repo.go     # ValidateWiFi
│   │   ├── gps_config_repo.go
│   │   └── shift_repo.go           # FindDefault
│   ├── usecase/
│   │   ├── attendance/main.go      # Anti-fraud + location validation
│   │   ├── user/main.go            # Login, bcrypt, JWT generation
│   │   ├── branch/main.go
│   │   └── report/main.go          # Dashboard stats + cache strategy
│   ├── handler/
│   │   ├── admin/                  # HTTP handlers cho Admin Portal
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── branch_handler.go
│   │   │   ├── attendance_handler.go
│   │   │   └── report_handler.go
│   │   └── user/                   # HTTP handlers cho Employee App
│   │       ├── auth_handler.go
│   │       └── attendance_handler.go
│   ├── middleware/
│   │   ├── auth.go                 # JWTAuth, RequireRole, GetUserID/Role/BranchID
│   │   ├── rate_limiter.go         # Redis sliding window limiter
│   │   └── request_logger.go       # slog structured logging + request ID
│   ├── infrastructure/
│   │   ├── cache/redis.go          # Redis client + Cache interface impl
│   │   ├── database/
│   │   │   ├── postgres.go         # GORM setup, connection pool
│   │   │   └── migrations/
│   │   │       └── migrations.go   # gormigrate: versioned migrations (entity embedding)
│   │   └── logger/setup.go         # slog JSON/Text setup
│   └── server/
│       └── router.go               # Echo route registration + middleware stack
├── pkg/
│   ├── apperrors/errors.go         # AppError struct, domain errors (ErrUserNotFound...)
│   ├── response/response.go        # OK, Created, Error, Paginated helpers
│   └── utils/
│       ├── jwt.go                  # GenerateToken, ParseToken, Claims
│       ├── geo.go                  # IsWithinGeofence (Haversine), IsValidCoordinate
│       └── pagination.go           # PaginationParams, ParsePaginationQuery
├── migrations/
│   └── 001_init_schema.sql         # Legacy SQL reference (replaced by gormigrate)
├── go.mod
└── go.sum
```

### Frontend (`sa-web/`)

```
sa-web/
├── app/
│   ├── layout.tsx                  # Root layout: font, metadata, Providers
│   ├── globals.css                 # CSS variables (light/dark) + Shimmer animation
│   ├── page.tsx                    # / → redirect /dashboard
│   ├── providers.tsx               # QueryClient + Toaster
│   ├── (auth)/
│   │   └── login/page.tsx          # Login form: Zod validation, JWT
│   └── (admin)/                    # Route group: layout với Sidebar
│       ├── layout.tsx              # Auth guard + Sidebar + Header
│       ├── dashboard/page.tsx      # KPI cards, Pie/Bar charts, branch table
│       ├── branches/page.tsx       # CRUD + GPS Geofencing config
│       ├── users/page.tsx          # List 5000 nhân viên + filters + CRUD
│       ├── attendance/page.tsx     # DataTable logs + fraud flags
│       └── reports/page.tsx        # Báo cáo NV/Chi nhánh + export
├── components/
│   ├── ui/                         # shadcn/ui primitives
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── dialog.tsx
│   │   ├── input.tsx
│   │   ├── label.tsx
│   │   ├── select.tsx
│   │   ├── skeleton.tsx            # Loading shimmer wrapper
│   │   ├── table.tsx
│   │   └── badge.tsx
│   ├── layout/
│   │   ├── sidebar.tsx             # Collapsible nav + active state
│   │   └── header.tsx              # Current user + logout
│   ├── shared/
│   │   ├── pagination.tsx          # Page-aware pagination controls
│   │   ├── status-badge.tsx        # StatusBadge, RoleBadge, ActiveBadge
│   │   └── data-table-skeleton.tsx # Table loading state
│   ├── branches/
│   │   └── branch-form-dialog.tsx  # GPS/WiFi config form
│   └── users/
│       ├── user-form-dialog.tsx    # Create/Edit user form
│       └── reset-password-dialog.tsx
├── hooks/                          # TanStack Query wrappers
│   ├── use-auth.ts                 # useCurrentUser, useLogin, useLogout
│   ├── use-users.ts                # useUsers, useCreateUser, useDeleteUser...
│   ├── use-branches.ts
│   ├── use-attendance.ts
│   └── use-reports.ts              # useDashboardStats (auto-refetch 5 phút)
├── services/                       # API service layer
│   ├── auth.service.ts             # login → set cookies + localStorage
│   ├── user.service.ts
│   ├── branch.service.ts
│   ├── attendance.service.ts
│   └── report.service.ts
├── lib/
│   ├── api-client.ts               # Axios instance + JWT interceptor + 401 handler
│   ├── auth.ts                     # isAuthenticated, getStoredUser, clearStoredUser
│   └── utils.ts                    # cn(), formatDate, formatTime, formatHours
├── types/                          # TypeScript interfaces (đồng bộ với Go structs)
│   ├── api.ts                      # ApiResponse<T>, PaginationMeta
│   ├── auth.ts                     # LoginRequest, LoginResponse, UserRole
│   ├── user.ts                     # User, CreateUserRequest, UserFilter
│   ├── branch.ts                   # Branch, WifiNetwork, CreateBranchRequest
│   ├── attendance.ts               # AttendanceLog, AttendanceStatus, CheckMethod
│   ├── report.ts                   # DashboardStats, BranchTodayStats
│   └── index.ts                    # Re-export tất cả
├── next.config.ts
├── tailwind.config.ts
├── tsconfig.json
├── components.json                 # shadcn/ui config
└── .env.local                      # NEXT_PUBLIC_API_URL
```

### Mobile (`sa-mb/`)

```
sa-mb/
├── lib/
│   ├── main.dart                   # Entry point
│   ├── app.dart                    # Root widget, DI, Routing
│   ├── core/
│   │   ├── constants/              # API URLs, app constants
│   │   ├── network/api_client.dart # Dio HTTP client + JWT interceptor
│   │   ├── theme/                  # HDBank colors, Material Design 3
│   │   └── utils/date_utils.dart
│   ├── data/
│   │   ├── models/                 # JSON ↔ Dart objects (khớp Go entity)
│   │   ├── repositories/           # API implementation (auth, attendance)
│   │   └── services/               # Platform services
│   │       ├── location_service.dart   # GPS: toạ độ, geofence, mock detect
│   │       ├── wifi_service.dart       # WiFi: SSID/BSSID scan
│   │       ├── device_service.dart     # device_id, model, app version
│   │       └── security_service.dart   # Anti-fraud: VPN + Fake GPS detect
│   ├── domain/
│   │   └── repositories/           # Abstract interfaces (Clean Architecture)
│   └── presentation/
│       ├── blocs/                  # BLoC state management (auth, attendance)
│       ├── screens/                # Login, Home, CheckIn, History
│       └── widgets/                # AttendanceCard, StatusBadge, LoadingOverlay
├── assets/                         # Images, icons
└── pubspec.yaml
```

---

## 8. Cài đặt & Chạy Project

### Yêu cầu môi trường

| Tool | Version | Mục đích |
|---|---|---|
| Go | ≥ 1.22 | Build backend API |
| Node.js | ≥ 20 | Build frontend (Admin Portal) |
| Flutter | ≥ 3.16 | Build mobile app (Employee App) |
| PostgreSQL | ≥ 16 | Primary database |
| Redis | ≥ 7 | Cache + rate limiting |
| Docker & Docker Compose | ≥ 24 | Container orchestration |

---

### Cách 1 — Docker Compose (Khuyến nghị)

Chạy toàn bộ hệ thống với một lệnh:

```bash
# 1. Clone repository
git clone https://github.com/hdbank/smart-attendance.git
cd smart-attendance

# 2. Copy và chỉnh sửa environment
cp sa-api/config/config.yaml.example sa-api/config/config.yaml
# Chỉnh sửa JWT_SECRET và các thông tin cần thiết

# 3. Build và khởi động tất cả services
docker compose up -d --build

# 4. Kiểm tra trạng thái
docker compose ps

# Kết quả mong đợi:
# NAME         STATUS    PORTS
# postgres     running   0.0.0.0:5432->5432/tcp
# redis        running   0.0.0.0:6379->6379/tcp
# sa-api       running   0.0.0.0:8080->8080/tcp
# sa-web       running   0.0.0.0:3000->3000/tcp

# 5. Kiểm tra API
curl http://localhost:8080/health
# → {"status":"ok","version":"1.0.0"}
```

**Truy cập:**
- Admin Portal: http://localhost:3000
- API: http://localhost:8080
- Tài khoản mặc định: `admin@hdbank.com.vn` / `Admin@123`

---

### Cách 2 — Chạy thủ công (Development)

#### Bước 1 — Chuẩn bị Database

```bash
# Tạo PostgreSQL database
psql -U postgres -c "CREATE DATABASE smart_attendance;"

# Migration sẽ tự động chạy khi server start (gormigrate).
# Hoặc chạy thủ công trước:
cd sa-api && go run ./cmd/migration -cmd up
```

#### Bước 2 — Cấu hình Backend

```bash
cd sa-api

# Copy config
cp config/config.yaml.example config/config.yaml

# Chỉnh sửa config/config.yaml
# database.host, database.password, redis.password, jwt.secret
```

```yaml
# config/config.yaml (development)
app:
  name: SmartAttendance
  port: "8080"
  env: development
  debug: true

database:
  host: localhost
  port: "5432"
  name: smart_attendance
  user: postgres
  password: your_password      # ← Thay đổi
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 300

redis:
  host: localhost
  port: "6379"
  password: ""                 # ← Thay đổi nếu Redis có password
  pool_size: 10

jwt:
  secret: "change-me-32chars-minimum" # ← BẮT BUỘC thay đổi
  expire_hours: 24
  refresh_expire_days: 7
```

#### Bước 3 — Khởi động Backend

```bash
cd sa-api

# Tải dependencies
go mod download

# Chạy development (hot reload với Air)
go install github.com/air-verse/air@latest
air

# Hoặc chạy trực tiếp
go run ./cmd/server

# Kiểm tra
curl http://localhost:8080/health
# → {"status":"ok"}
```

Logs mong đợi:
```
level=INFO msg="database connected" host=localhost db=smart_attendance
level=INFO msg="running database migrations..."
level=INFO msg="database migrations completed"
level=INFO msg="redis connected" host=localhost
level=INFO msg="server listening" addr=:8080
```

#### Bước 4 — Khởi động Frontend

```bash
cd sa-web

# Tải dependencies
npm install

# Copy và cấu hình environment
cp .env.local.example .env.local
# NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1

# Khởi động development server
npm run dev

# Truy cập: http://localhost:3000
```

#### Bước 5 — Khởi động Mobile App

```bash
cd sa-mb

# Sinh platform folders (chỉ lần đầu)
flutter create .

# Cài dependencies
flutter pub get

# Chạy trên emulator/simulator
flutter run
```

> Cấu hình API URL tại `lib/core/constants/api_constants.dart`:
> - Android Emulator: `http://10.0.2.2:8080`
> - iOS Simulator: `http://localhost:8080`
> - Thiết bị thật: `http://<IP-LAN>:8080`

#### Bước 6 — Xác nhận hệ thống hoạt động

```bash
# Test Login API
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@hdbank.com.vn","password":"Admin@123"}'

# Response mong đợi:
# {
#   "success": true,
#   "data": {
#     "access_token": "eyJhbGci...",
#     "refresh_token": "eyJhbGci...",
#     "user": {"id":1,"name":"System Administrator","role":"admin",...}
#   }
# }

# Test Admin Dashboard (dùng token vừa lấy)
curl http://localhost:8080/api/v1/admin/reports/dashboard \
  -H "Authorization: Bearer eyJhbGci..."
```

---

### Production Checklist

Trước khi deploy production, kiểm tra các mục sau:

```bash
# ✅ Bảo mật
[ ] JWT_SECRET tối thiểu 32 ký tự ngẫu nhiên
[ ] DATABASE_PASSWORD mạnh, không phải "postgres"
[ ] config.yaml không chứa secret thật (dùng env vars)
[ ] SSL/TLS được bật cho PostgreSQL (sslmode=require)
[ ] CORS origin được giới hạn (không dùng *)

# ✅ Database
[ ] Chạy migration trên production DB
[ ] Đã tạo đầy đủ indexes (kiểm tra bằng \d table_name)
[ ] Backup strategy đã được thiết lập

# ✅ Hiệu năng
[ ] max_open_conns phù hợp với số lượng users
[ ] Redis maxmemory được set (ví dụ: 512mb)
[ ] Rate limit thresholds phù hợp với traffic thực tế

# ✅ Monitoring
[ ] Health check endpoint /health được expose
[ ] Log level = INFO (không debug) trên production
[ ] Alert cho error rate > 1%
```

---

### Biến môi trường quan trọng

| Biến | Ví dụ | Mô tả |
|---|---|---|
| `DATABASE_HOST` | `postgres` | PostgreSQL host |
| `DATABASE_PASSWORD` | `StrongPass!123` | **BẮT BUỘC thay đổi** |
| `DATABASE_MAX_OPEN_CONNS` | `25` | Connection pool size |
| `REDIS_HOST` | `redis` | Redis host |
| `REDIS_PASSWORD` | `RedisPass!` | Redis auth |
| `JWT_SECRET` | `random-32-char-string` | **BẮT BUỘC thay đổi** |
| `APP_ENV` | `production` | `development` / `production` |
| `APP_DEBUG` | `false` | Tắt debug log trên production |
| `NEXT_PUBLIC_API_URL` | `https://api.hdbank.vn/api/v1` | Backend URL cho frontend |

---

*Cập nhật lần cuối: 2026-03-30*
