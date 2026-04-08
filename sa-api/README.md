# Smart Attendance API

Hệ thống **Chấm công thông minh** (Smart Attendance) cho doanh nghiệp quy mô **100 chi nhánh** và **5.000 nhân viên**, xây dựng theo mô hình **Clean Architecture** với Golang.

---

## Mục lục

- [Tổng quan kiến trúc](#tổng-quan-kiến-trúc)
- [Cấu trúc source code](#cấu-trúc-source-code)
- [Tính năng đang triển khai](#tính-năng-đang-triển-khai)
- [Tính năng có thể phát triển thêm](#tính-năng-có-thể-phát-triển-thêm)
- [Yêu cầu hệ thống](#yêu-cầu-hệ-thống)
- [Hướng dẫn cài đặt và chạy](#hướng-dẫn-cài-đặt-và-chạy)
- [Database Migration](#database-migration)
- [API Reference](#api-reference)
- [Thiết kế Database](#thiết-kế-database)
- [Bảo mật và Anti-Fraud](#bảo-mật-và-anti-fraud)
- [Tối ưu hiệu năng](#tối-ưu-hiệu-năng)

---

## Tổng quan kiến trúc

Dự án áp dụng **Clean Architecture** (còn gọi là Onion Architecture) với 4 layer rõ ràng:

```
┌─────────────────────────────────────────┐
│           Handler Layer (HTTP)          │  ← Nhận request, validate input, RBAC
├─────────────────────────────────────────┤
│         Usecase Layer (Business)        │  ← Business logic, anti-fraud, location
├─────────────────────────────────────────┤
│       Repository Layer (Data Access)    │  ← Truy vấn DB/Cache
├─────────────────────────────────────────┤
│         Domain Layer (Entities)         │  ← Entity, Interface contracts
└─────────────────────────────────────────┘
```

**Nguyên tắc thiết kế:**
- **Dependency Inversion**: Usecase chỉ phụ thuộc vào Interface, không phụ thuộc trực tiếp vào PostgreSQL/Redis
- **SOLID**: Mỗi struct có một trách nhiệm duy nhất, dễ mở rộng và test
- **Interface-driven**: Tất cả Repository và Usecase đều được định nghĩa qua Interface → dễ mock để Unit Test
- **RBAC 3 cấp**: Admin (toàn hệ thống) → Manager (chi nhánh) → Employee (cá nhân)

---

## Cấu trúc source code

```
sa-api/
├── cmd/
│   ├── server/
│   │   └── main.go                        # Entry point, Dependency Injection, auto-migration
│   └── migration/
│       └── main.go                        # CLI tool: migrate / rollback độc lập
│
├── config/
│   ├── config.go                          # Viper loader, struct Config
│   └── config.yaml                        # Giá trị config (dev). KHÔNG commit secret thật
│
├── internal/
│   ├── domain/                            # Layer trung tâm — không import package nào khác trong project
│   │   ├── entity/
│   │   │   ├── branch.go                  # Chi nhánh (GPS coords, city/province)
│   │   │   ├── user.go                    # Người dùng — RBAC 3 cấp (Admin/Manager/Employee)
│   │   │   ├── attendance.go              # Log chấm công (UNIQUE user+date)
│   │   │   ├── daily_summary.go           # Pre-computed aggregate per chi nhánh/ngày
│   │   │   ├── wifi_config.go             # Cấu hình WiFi SSID/BSSID cho phép
│   │   │   ├── gps_config.go              # Cấu hình GPS Geofencing (bán kính mét)
│   │   │   ├── shift.go                   # Ca làm việc
│   │   │   ├── correction.go             # Yêu cầu bổ sung công (ca chính thức + tăng ca)
│   │   │   ├── leave.go                  # Yêu cầu nghỉ phép
│   │   │   └── overtime.go               # Yêu cầu tăng ca (OT)
│   │   ├── repository/                    # Interfaces cho data access
│   │   │   ├── attendance.go              # AttendanceFilter, BranchTodayStats, TodayEmployeeFilter
│   │   │   ├── branch.go
│   │   │   ├── user.go
│   │   │   ├── wifi_config.go
│   │   │   ├── gps_config.go
│   │   │   ├── shift.go
│   │   │   ├── correction.go             # CorrectionFilter, CountByUserInMonth
│   │   │   ├── leave.go                  # LeaveFilter
│   │   │   └── overtime.go               # OvertimeFilter
│   │   └── usecase/                       # Interfaces + request/response types
│   │       ├── attendance.go              # CheckInRequest, CheckOutRequest
│   │       ├── user.go                    # LoginRequest/Response, CreateUserRequest
│   │       ├── branch.go
│   │       ├── report.go                  # TodayStatsFilter, DashboardStats, ReportFilter
│   │       ├── correction.go             # CreateCorrectionRequest, ProcessCorrectionRequest
│   │       ├── leave.go                  # CreateLeaveRequest, PendingApprovalItem
│   │       └── overtime.go               # OvertimeCheckInRequest, ProcessOvertimeRequest
│   │
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── postgres.go                # GORM setup + connection pool tối ưu
│   │   │   └── migrations/
│   │   │       └── migrations.go          # Danh sách migration versioned (gormigrate)
│   │   ├── cache/
│   │   │   └── redis.go                   # Redis client với Cache interface
│   │   └── logger/
│   │       └── logger.go                  # slog setup (JSON/Text theo môi trường)
│   │
│   ├── repository/                        # Triển khai Repository interfaces (PostgreSQL)
│   │   ├── attendance.go                  # CRUD + CTE aggregate + GetTodayStatsByBranch
│   │   ├── branch.go
│   │   ├── user.go
│   │   ├── wifi_config.go
│   │   ├── gps_config.go
│   │   ├── shift.go
│   │   └── helper.go
│   │
│   ├── usecase/                           # Triển khai Usecase (Business Logic)
│   │   ├── attendance/
│   │   │   └── main.go                    # Check-in/out + WiFi/GPS + anti-fraud + cache
│   │   ├── user/
│   │   │   └── main.go                    # Login + JWT + quản lý nhân viên
│   │   ├── branch/
│   │   │   └── main.go                    # Quản lý chi nhánh
│   │   ├── report/
│   │   │   └── main.go                    # Dashboard + today stats + báo cáo
│   │   ├── correction/
│   │   │   └── main.go                    # Bổ sung công (ca chính thức + tăng ca) + phê duyệt
│   │   ├── leave/
│   │   │   └── main.go                    # Nghỉ phép + phê duyệt tổng hợp
│   │   └── overtime/
│   │       └── main.go                    # Tăng ca: check-in/out + bo tròn + phê duyệt
│   │
│   ├── handler/
│   │   ├── admin/                         # HTTP Handlers cho Admin Portal
│   │   │   ├── auth.go                    # Login (403 cho employee), Me, ChangePassword
│   │   │   ├── user.go                    # CRUD nhân viên
│   │   │   ├── branch.go                  # CRUD chi nhánh
│   │   │   ├── attendance.go              # GetList, GetByID, GetSummary
│   │   │   ├── report.go                  # Dashboard, Today stats, Today employees, Reports
│   │   │   ├── correction.go             # Duyệt bổ sung công
│   │   │   ├── leave.go                  # Duyệt nghỉ phép + unified approvals
│   │   │   ├── overtime.go               # Duyệt tăng ca
│   │   │   └── wifi_config.go            # CRUD cấu hình WiFi
│   │   └── user/                          # HTTP Handlers cho Employee App
│   │       ├── auth.go                    # Login, Me, ChangePassword
│   │       ├── attendance.go              # CheckIn, CheckOut, GetMyToday
│   │       ├── correction.go             # Tạo yêu cầu bổ sung công
│   │       ├── leave.go                  # Tạo yêu cầu nghỉ phép
│   │       ├── overtime.go               # Check-in/out tăng ca
│   │       └── helper.go                  # getUserIDFromContext()
│   │
│   ├── middleware/
│   │   ├── auth.go                        # JWTAuth, RequireRole, GetUserID/BranchID/Role
│   │   ├── rate_limiter.go                # Rate limiting per IP / per user (Redis)
│   │   └── request_logger.go             # Structured request logging với slog
│   │
│   └── server/
│       └── router.go                      # Echo route registration — tất cả routes
│
├── pkg/
│   ├── apperrors/
│   │   └── errors.go                      # AppError, domain errors, ValidationError
│   ├── response/
│   │   └── response.go                    # OK, Created, Paginated, Error — chuẩn hóa JSON
│   └── utils/
│       ├── geo.go                         # Haversine formula, GPS geofencing
│       ├── jwt.go                         # GenerateToken / ParseToken
│       └── pagination.go                  # ParsePagination, PaginationParams
│
├── migrations/
│   └── 001_init_schema.sql                # Legacy SQL schema (tham khảo)
│
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

---

## Tính năng đang triển khai

### 1. Phân quyền RBAC 3 cấp

| Role | Phạm vi | Quyền |
|------|---------|-------|
| **Admin** | Toàn hệ thống | Tất cả chi nhánh, tạo/xóa chi nhánh, quản lý mọi nhân viên |
| **Manager** | Chi nhánh của mình | Xem/sửa nhân viên, báo cáo chi nhánh mình — không thấy chi nhánh khác |
| **Employee** | Cá nhân | Chỉ check-in/check-out và xem thông tin bản thân |

> **Quan trọng về bảo mật**: `BranchID` khi check-in luôn được lấy từ profile DB của user — không tin vào giá trị client gửi lên. Manager bị lock vào chi nhánh mình ở mọi API.

### 2. Xác thực vị trí check-in/out (Dual Validation)

```
WiFi match?  ──YES──► Cho phép check-in
    │
    NO
    │
GPS trong geofence? ──YES──► Cho phép check-in
    │
    NO
    │
Từ chối — LOCATION_INVALID
```

- **WiFi**: So khớp SSID + BSSID (MAC address router). Một chi nhánh có nhiều WiFi config.
- **GPS Geofencing**: Tính khoảng cách Haversine, cho phép nếu trong bán kính cấu hình (mặc định 100m). Một chi nhánh có nhiều vùng geofence.
- **WiFi ưu tiên** trước GPS — chính xác hơn trong tòa nhà, tránh GPS drift.

### 3. Hệ thống chống gian lận (Anti-Fraud)

Luồng kiểm tra trước mỗi check-in:

```
[1] Fake GPS flag (từ mobile SDK)      ──vi phạm──► 403 FAKE_GPS_DETECTED
[2] VPN flag (từ mobile SDK)           ──vi phạm──► 403 VPN_DETECTED
[3] IP Blocklist (Redis: blocked:ip:*) ──vi phạm──► 403 VPN_DETECTED
[4] Suspicious count ≥ 3 / 7 ngày     ──vi phạm──► 403 TOO_MANY_SUSPICIOUS
[5] Đã check-in hôm nay chưa?         ──có rồi──► 409 ALREADY_CHECKED_IN
[6] Validate Location (WiFi → GPS)    ──ngoài vùng──► 403 LOCATION_INVALID
```

- **IP Blocklist**: Admin quản lý danh sách IP bị chặn qua Redis key `blocked:ip:{ip}`. Fail-open — nếu Redis lỗi, vẫn cho phép request.
- **Device ID Tracking**: Ghi lại device_id mỗi lần chấm công để phát hiện dùng chung thiết bị.

### 4. Tính toán giờ làm & trạng thái

| Trạng thái | Điều kiện |
|------------|-----------|
| `present` | Check-in trước `start_time + late_after` phút |
| `late` | Check-in sau `start_time + late_after` phút |
| `early_leave` | Check-out trước `end_time - early_before` phút |
| `half_day` | Tổng giờ làm < 50% `work_hours` của ca |
| `absent` | Không có bản ghi chấm công trong ngày |

### 5. Bổ sung công (Chấm công bù)

Hỗ trợ 2 loại bổ sung công qua `correction_type`:

#### Ca chính thức (`attendance`)
- Nhân viên đăng ký bù cho ngày bị trễ/về sớm
- Khi Manager duyệt → cập nhật check-in/out về giờ chuẩn của ca
- Hạn mức: tối đa 4 credits/tháng (cấu hình `correction.max_per_month`)

#### Tăng ca (`overtime`)
- Nhân viên đăng ký bổ sung khi quên check-in hoặc check-out tăng ca
- Khi tạo yêu cầu → tự động bổ sung thời gian thiếu (18:00 cho check-in, 22:00 cho check-out) và chuyển OT sang `pending`
- Khi Manager duyệt → tính bo tròn và approve OT trong cùng transaction
- Hạn mức riêng: tối đa 4 credits/tháng (cấu hình `correction.overtime_max_per_month`)

### 6. Nghỉ phép

- Employee đăng ký nghỉ phép (full_day, half_day_morning, half_day_afternoon)
- Ngày quá khứ: tự detect absent → full_day, half_day → nửa ngày còn lại
- Ngày tương lai: chọn loại nghỉ tự do
- Kiểm tra số ngày phép còn lại trước khi tạo và khi duyệt
- Khi duyệt → tạo/cập nhật attendance_log với status=leave trong transaction
- Auto-reject yêu cầu PENDING tháng cũ vào ngày 1 hàng tháng
- Cộng 1 ngày phép/tháng cho tất cả user active (scheduler)

### 7. Tăng ca (Overtime)

- Check-in OT chỉ được phép sau 17:00
- Logic "Bo tròn" thời gian:
  - Check-in trong [17:00-18:00] → tính từ 18:00; sau 18:00 → tính theo thực tế
  - Check-out sau 22:00 → tính đến 22:00; trước 22:00 → tính theo thực tế
- Thời gian tối đa: 4 giờ/ngày (18:00 - 22:00)
- Status flow: `init` (thiếu check-in hoặc check-out) → `pending` (đủ cả hai) → `approved`/`rejected`
- Hỗ trợ check-out mà không có check-in (quên check-in) → status `init`, cần bổ sung công
- Giờ OT chỉ cộng vào quỹ lương sau khi Manager Approve
- Auto-reject yêu cầu `init`/`pending` tháng cũ

### 8. Phê duyệt tổng hợp

- API unified `GET /admin/approvals` merge cả 3 loại: bổ sung công + nghỉ phép + tăng ca
- Hỗ trợ filter theo status (pending/approved/rejected)
- Sort theo created_at DESC, phân trang
- Manager chỉ thấy yêu cầu của chi nhánh mình, không được self-approve
- Batch approve cho tất cả yêu cầu pending

### 9. Soft Delete

- Tất cả bảng đều có `deleted_at` (sử dụng `gorm.DeletedAt`)
- GORM tự động thêm `WHERE deleted_at IS NULL` vào mọi query
- Unique indexes có `WHERE deleted_at IS NULL` để cho phép recreate sau khi soft-delete

### 10. Dashboard & Báo cáo

#### API thống kê hôm nay (`GET /api/v1/admin/reports/today`)

Trả về số liệu tổng hợp **per chi nhánh** trong ngày hôm nay. Sử dụng CTE SQL tính song song, không N+1:

```json
{
  "branch_id": 1,
  "branch_name": "Chi nhánh Hà Nội",
  "total_employees": 50,
  "present_count": 30,
  "late_count": 8,
  "early_leave_count": 2,
  "half_day_count": 1,
  "absent_count": 9,
  "suspicious_count": 3,
  "attendance_rate": 82.00,
  "on_time_rate": 73.17
}
```

#### API danh sách nhân viên hôm nay (`GET /api/v1/admin/reports/today/employees`)

Danh sách chi tiết từng nhân viên với derived status. **Đặc biệt**: trả về cả nhân viên `absent` (chưa có record trong DB) bằng LEFT JOIN từ `users`.

Filter `status`:
- `present` / `late` / `early_leave` / `half_day` → attendance_logs.status gốc
- `absent` → nhân viên không có bản ghi hôm nay
- `suspicious` → bất kỳ bản ghi nào có `is_fake_gps=true` hoặc `is_vpn=true`

#### Cache strategy cho dashboard

| API | TTL | Ghi chú |
|-----|-----|---------|
| `today/employees` | 2 phút | Near-realtime |
| `today` (branch stats) | 2 phút | Near-realtime |
| `dashboard` (admin) | 5 phút | Aggregate nặng |
| `dashboard` (branch) | 5 phút | Aggregate theo branch |

### 6. Caching với Redis

| Dữ liệu | Cache key | TTL |
|---------|-----------|-----|
| User info | `user:{id}` | 10 phút |
| Chi nhánh active | `branch:active` | 15 phút |
| Check-in hôm nay | `attend:today:{user_id}` | 5 phút |
| Dashboard admin | `dashboard:admin:stats` | 5 phút |
| Today branch stats | `dashboard:today:branch:...` | 2 phút |
| Today employees | `dashboard:today:employees:...` | 2 phút |
| IP blocklist | `blocked:ip:{ip}` | Admin-managed |

> Sau mỗi check-in/check-out, cache `SET` ngay giá trị mới (không chỉ `DELETE`) để request tiếp theo hit cache thay vì DB.

### 7. Bảo mật

- **JWT HS256**: Access token 24h, Refresh token 7 ngày
- **Rate limiting**: 10 lần login/15 phút per IP; 10 lần check-in/phút per user
- **Bcrypt** cost 10 cho password
- **Fail-open rate limiter**: Redis lỗi → vẫn cho request qua, không block service
- **Graceful shutdown**: Chờ request đang xử lý hoàn thành (timeout 10s)
- **Admin Portal login riêng**: `POST /admin/auth/login` trả về `403 FORBIDDEN` nếu role là `employee`

---

## Tính năng có thể phát triển thêm

### Nghiệp vụ
- [x] ~~**Đơn xin phép / Nghỉ phép**: Module leave request với approval workflow~~ ✅ Đã triển khai
- [x] ~~**Bổ sung công (Chấm công bù)**: Bù công cho ca chính thức và tăng ca~~ ✅ Đã triển khai
- [x] ~~**Tăng ca (Overtime)**: Check-in/out OT, bo tròn 18:00-22:00, phê duyệt~~ ✅ Đã triển khai
- [ ] **Chấm công khuôn mặt (Face Recognition)**: Tích hợp AI xác thực qua camera
- [ ] **Push Notification**: Nhắc nhở check-in/out qua Firebase
- [ ] **Export báo cáo**: Xuất Excel/PDF cho HR/Payroll
- [ ] **QR Code Check-in**: QR code động thay thế GPS trong văn phòng
- [ ] **Webhook Events**: Gửi event check-in/out tới hệ thống HR/Payroll
- [ ] **Daily Summary Job**: Scheduler tính toán và upsert `daily_summaries` cuối ngày
- [ ] **Quản lý ngày lễ**: Cấu hình ngày lễ theo năm, tự động tính công

### Kỹ thuật
- [ ] **WebSocket Real-time**: Cập nhật dashboard live khi có nhân viên check-in
- [ ] **Message Queue (Kafka/RabbitMQ)**: Xử lý audit log bất đồng bộ
- [ ] **Prometheus + Grafana**: Metrics dashboard cho DevOps
- [ ] **Distributed Tracing (Jaeger/Zipkin)**: Trace request end-to-end
- [ ] **Swagger/OpenAPI**: Auto-generate API documentation
- [ ] **Unit/Integration Tests**: Mock repository layer

---

## Yêu cầu hệ thống

| Phần mềm | Phiên bản |
|----------|-----------|
| Go | 1.22+ |
| PostgreSQL | 15+ |
| Redis | 7+ |
| Docker & Docker Compose | (tùy chọn, khuyến nghị) |

---

## Hướng dẫn cài đặt và chạy

### Cách 1: Docker Compose (Khuyến nghị)

```bash
git clone <repository-url>
cd sa-api

docker-compose up -d

# Kiểm tra
curl http://localhost:8080/health
# {"status":"ok","service":"smart-attendance"}
```

### Cách 2: Chạy thủ công (Development)

```bash
# Khởi động PostgreSQL + Redis
docker-compose up -d postgres redis

# Cài dependencies
go mod download

# Chạy server (migration tự động chạy khi start)
go run ./cmd/server
```

### Tài khoản mặc định (seed bởi migration `20250330000001`)

| Field | Giá trị |
|-------|---------|
| Email | `admin@hdbank.com.vn` |
| Password | `Admin@123` |
| Role | `admin` |

---

## Database Migration

Dự án dùng **[gormigrate](https://github.com/go-gormigrate/gormigrate)** để quản lý schema versioned. Mỗi migration có ID dạng timestamp (`YYYYMMDDHHMMSS`), được track trong bảng `migrations` của DB — chỉ migration chưa có trong bảng mới được thực thi.

### Cấu trúc file

```
sa-api/
├── cmd/migration/
│   └── main.go                                    # CLI tool: -cmd up|down|rollback-to|reset
└── internal/infrastructure/database/migrations/
    └── migrations.go                              # GetMigrations() — danh sách migration
```

Migration định nghĩa struct bằng cách **embed entity** hiện tại (ví dụ `entity.User`), phù hợp vì GORM AutoMigrate chỉ thêm cột mới — không bao giờ xóa cột cũ:

```go
// internal/infrastructure/database/migrations/migrations.go

{
    ID: "20250330000001",
    Migrate: func(tx *gorm.DB) error {
        type User struct {
            entity.User  // Embed entity trực tiếp
        }
        return tx.AutoMigrate(&User{})
    },
    Rollback: func(tx *gorm.DB) error {
        return tx.Migrator().DropTable("users")
    },
},
```

### Tự động chạy khi server start

Khi `cmd/server/main.go` khởi động, gormigrate được gọi trước khi server nhận request:

```go
m := gormigrate.New(db, gormigrate.DefaultOptions, migrations.GetMigrations())
m.Migrate()  // chỉ apply migration chưa có trong bảng "migrations"
```

```
Server start
    │
    ▼
gormigrate.Migrate()
    │
    ├── 20250330000001 đã chạy? → skip
    ├── 20250330000002 đã chạy? → skip
    ├── 20250330000003 chưa có  → EXECUTE ✓
    └── Done — server tiếp tục khởi động
```

### CLI tool độc lập

```bash
# Migrate tất cả (default)
go run ./cmd/migration -cmd up

# Rollback migration cuối cùng
go run ./cmd/migration -cmd down

# Rollback về một ID cụ thể
go run ./cmd/migration -cmd rollback-to -id 20250330000001

# Reset toàn bộ (rollback tất cả migration)
go run ./cmd/migration -cmd reset
```

### Danh sách migration hiện tại

| ID | Nội dung |
|----|---------|
| `20250330000001` | Tạo 6 bảng core (branches, users, wifi_configs, gps_configs, shifts, attendance_logs) + seed admin |
| `20250330000002` | Tạo bảng `daily_summaries` |
| `20250330000003` | Thêm partial indexes (fraud detection, BSSID lookup) |
| `20250330000004` | Tạo bảng `attendance_corrections` (bổ sung công) |
| `20250331000001` | Remove latitude/longitude from branches (moved to gps_configs) |
| `20250406000001` | Thêm `credit_count` vào attendance_corrections |
| `20250407000001` | Tạo bảng `leave_requests` (nghỉ phép) |
| `20250407000002` | Thêm `leave_balance` vào users |
| `20250407000003` | Bổ sung index hiệu năng cho leave/correction |
| `20250407000004` | Tạo bảng `overtime_requests` (tăng ca) |
| `20250407000005` | Thêm `deleted_at` vào tất cả bảng + bỏ column overtime + thêm overtime_request_id |
| `20250407000006` | Thêm `correction_type`, `overtime_request_id` cho corrections + partial unique indexes |

### Quy ước khi thêm migration mới

```go
// File: internal/infrastructure/database/migrations/migrations.go
// Thêm block mới vào cuối slice trong GetMigrations():

{
    ID: "20260101120000",
    Migrate: func(tx *gorm.DB) error {
        type MyNewTable struct {
            entity.MyNewEntity
        }
        return tx.AutoMigrate(&MyNewTable{})
    },
    Rollback: func(tx *gorm.DB) error {
        return tx.Migrator().DropTable("my_new_tables")
    },
},
```

---

## API Reference

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
Tất cả API (trừ login) yêu cầu header:
```
Authorization: Bearer <access_token>
```

### Employee App — `/api/v1/auth/*` và `/api/v1/attendance/*`

#### Auth
| Method | Path | Mô tả | Auth |
|--------|------|-------|------|
| `POST` | `/auth/login` | Đăng nhập (tất cả role) | Không |
| `GET` | `/auth/me` | Thông tin user hiện tại | JWT |
| `PUT` | `/auth/change-password` | Đổi mật khẩu | JWT |

#### Attendance (Employee App)
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `POST` | `/attendance/check-in` | Check-in chấm công | All |
| `POST` | `/attendance/check-out` | Check-out kết thúc ca | All |
| `GET` | `/attendance/today` | Trạng thái hôm nay | All |
| `GET` | `/attendance/history` | Lịch sử chấm công | All |

#### Corrections — Bổ sung công (Employee App)
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `POST` | `/attendance/corrections` | Tạo yêu cầu bổ sung công (attendance hoặc overtime) | All |
| `GET` | `/attendance/corrections` | Danh sách yêu cầu của bản thân | All |
| `GET` | `/attendance/corrections/:id` | Chi tiết yêu cầu | All |

#### Leaves — Nghỉ phép (Employee App)
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `POST` | `/attendance/leaves` | Tạo yêu cầu nghỉ phép | All |
| `GET` | `/attendance/leaves` | Danh sách yêu cầu của bản thân | All |
| `GET` | `/attendance/leaves/:id` | Chi tiết yêu cầu | All |

#### Overtime — Tăng ca (Employee App)
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `POST` | `/attendance/overtime/check-in` | Check-in tăng ca (chỉ sau 17:00) | All |
| `POST` | `/attendance/overtime/check-out` | Check-out tăng ca | All |
| `GET` | `/attendance/overtime/today` | Trạng thái OT hôm nay | All |
| `GET` | `/attendance/overtime` | Lịch sử tăng ca | All |
| `GET` | `/attendance/overtime/:id` | Chi tiết yêu cầu OT | All |

**Request check-in:**
```json
{
  "latitude": 10.7769,
  "longitude": 106.7009,
  "ssid": "HDB-Office-5G",
  "bssid": "AA:BB:CC:DD:EE:FF",
  "device_id": "unique-device-id",
  "device_model": "iPhone 15",
  "app_version": "1.0.0",
  "is_fake_gps": false,
  "is_vpn": false
}
```

> `branch_id` **không cần** gửi từ client — server tự lấy từ profile user trong DB.

---

### Admin Portal — `/api/v1/admin/*`

#### Admin Auth
| Method | Path | Mô tả | Auth |
|--------|------|-------|------|
| `POST` | `/admin/auth/login` | Đăng nhập Admin Portal (403 nếu role = employee) | Không |
| `GET` | `/admin/auth/me` | Thông tin admin/manager hiện tại | JWT |
| `PUT` | `/admin/auth/change-password` | Đổi mật khẩu | JWT |

#### Users
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `GET` | `/admin/users` | Danh sách nhân viên | Admin, Manager |
| `POST` | `/admin/users` | Tạo nhân viên mới | Admin, Manager |
| `GET` | `/admin/users/:id` | Chi tiết nhân viên | Admin, Manager |
| `PUT` | `/admin/users/:id` | Cập nhật thông tin | Admin, Manager |
| `DELETE` | `/admin/users/:id` | Vô hiệu hóa | Admin |
| `POST` | `/admin/users/:id/reset-password` | Đặt lại mật khẩu | Admin, Manager |

#### Branches
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `GET` | `/admin/branches` | Danh sách chi nhánh | Admin, Manager |
| `GET` | `/admin/branches/active` | Chi nhánh đang hoạt động | All |
| `POST` | `/admin/branches` | Tạo chi nhánh mới | Admin |
| `GET` | `/admin/branches/:id` | Chi tiết chi nhánh | Admin, Manager |
| `PUT` | `/admin/branches/:id` | Cập nhật chi nhánh | Admin |
| `DELETE` | `/admin/branches/:id` | Vô hiệu hóa chi nhánh | Admin |

#### Attendance (Admin)
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `GET` | `/admin/attendance` | Danh sách log chấm công (filter + phân trang) | Admin, Manager |
| `GET` | `/admin/attendance/:id` | Chi tiết bản ghi | Admin, Manager |
| `GET` | `/admin/attendance/summary/:user_id` | Thống kê theo nhân viên | Admin, Manager |

#### Corrections — Bổ sung công (Admin)
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `GET` | `/admin/corrections` | Danh sách yêu cầu bổ sung công | Admin, Manager |
| `GET` | `/admin/corrections/:id` | Chi tiết yêu cầu | Admin, Manager |
| `PUT` | `/admin/corrections/:id/process` | Duyệt/từ chối yêu cầu | Admin, Manager |
| `POST` | `/admin/corrections/batch-approve` | Duyệt tất cả pending | Admin, Manager |

#### Leaves — Nghỉ phép (Admin)
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `GET` | `/admin/leaves` | Danh sách yêu cầu nghỉ phép | Admin, Manager |
| `GET` | `/admin/leaves/:id` | Chi tiết yêu cầu | Admin, Manager |
| `PUT` | `/admin/leaves/:id/process` | Duyệt/từ chối yêu cầu | Admin, Manager |
| `POST` | `/admin/leaves/batch-approve` | Duyệt tất cả pending | Admin, Manager |

#### Overtime — Tăng ca (Admin)
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `GET` | `/admin/overtime` | Danh sách yêu cầu tăng ca | Admin, Manager |
| `GET` | `/admin/overtime/:id` | Chi tiết yêu cầu | Admin, Manager |
| `PUT` | `/admin/overtime/:id/process` | Duyệt/từ chối yêu cầu | Admin, Manager |
| `POST` | `/admin/overtime/batch-approve` | Duyệt tất cả pending | Admin, Manager |

#### Phê duyệt tổng hợp
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `GET` | `/admin/approvals` | Danh sách tổng hợp (bổ sung công + nghỉ phép + tăng ca) | Admin, Manager |
| `GET` | `/admin/approvals/pending` | Chỉ yêu cầu đang chờ duyệt | Admin, Manager |

#### Reports & Dashboard
| Method | Path | Mô tả | Role |
|--------|------|-------|------|
| `GET` | `/admin/reports/today` | **Thống kê hôm nay per chi nhánh** (phân trang) | Admin, Manager |
| `GET` | `/admin/reports/today/employees` | **Danh sách nhân viên hôm nay** theo status (phân trang) | Admin, Manager |
| `GET` | `/admin/reports/dashboard` | Dashboard tổng quan | Admin, Manager |
| `GET` | `/admin/reports/attendance` | Báo cáo chấm công theo kỳ | Admin, Manager |
| `GET` | `/admin/reports/branches` | Báo cáo tổng hợp per chi nhánh | Admin |
| `GET` | `/admin/reports/users/:user_id` | Báo cáo nhân viên cụ thể | Admin, Manager |

**Query params — `/admin/reports/today`:**
| Param | Mô tả |
|-------|-------|
| `branch_id` | Filter theo chi nhánh (chỉ admin tổng; manager bị lock vào chi nhánh mình) |
| `page`, `limit` | Phân trang |

**Query params — `/admin/reports/today/employees`:**
| Param | Giá trị |
|-------|---------|
| `status` | `present` / `late` / `early_leave` / `half_day` / `absent` / `suspicious` / bỏ trống = tất cả |
| `branch_id` | Filter theo chi nhánh (chỉ admin tổng) |
| `page`, `limit` | Phân trang |

**RBAC cho tất cả Reports API:**

| Caller | Hành vi |
|--------|---------|
| **Admin** | Không truyền `branch_id` → tất cả chi nhánh. Truyền `branch_id` → filter theo chi nhánh đó |
| **Manager** | Luôn chỉ thấy chi nhánh của mình; `branch_id` từ query bị bỏ qua |

### Response format chuẩn

```json
// Thành công
{
  "success": true,
  "data": { ... }
}

// Danh sách có phân trang
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}

// Lỗi
{
  "success": false,
  "error": {
    "code": "LOCATION_INVALID",
    "message": "Vị trí không hợp lệ"
  }
}

// Lỗi validation
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Dữ liệu không hợp lệ",
    "fields": {
      "email": "Email không được để trống",
      "password": "Mật khẩu phải có ít nhất 6 ký tự"
    }
  }
}
```

---

## Thiết kế Database

```
branches
├── id, code (UNIQUE), name, address, city, province, phone, email
├── latitude, longitude          ← Tọa độ trụ sở (hiển thị map)
└── is_active

users
├── id, employee_code (UNIQUE), email (UNIQUE)
├── branch_id → branches.id      ← NULL chỉ dành cho admin
├── name, phone, password (bcrypt), department, position
├── role: admin | manager | employee
├── hired_at (DATE)
└── is_active, last_login_at

shifts
├── id, branch_id → branches.id
├── name, start_time (HH:MM), end_time (HH:MM)
├── late_after (phút), early_before (phút), work_hours
└── is_default, is_active

wifi_configs
├── id, branch_id → branches.id
├── ssid, bssid (MAC router)
└── is_active

gps_configs
├── id, branch_id → branches.id
├── latitude, longitude, radius (mét)
└── is_active

attendance_logs            ← ~1.8M rows/năm với 5000 nhân viên
├── id, user_id, branch_id, shift_id, date
├── check_in_time, check_in_lat/lng, check_in_method (wifi|gps)
├── check_in_ssid, check_in_bssid
├── check_out_time, check_out_lat/lng, check_out_method
├── device_id, device_model, ip_address, app_version
├── is_fake_gps, is_vpn, fraud_note    ← Anti-fraud
├── status: present|late|early_leave|absent|half_day|leave|half_day_leave
├── work_hours, overtime_request_id → overtime_requests.id
├── deleted_at                          ← Soft delete
└── UNIQUE(user_id, date)              ← Một user một bản ghi/ngày

attendance_corrections     ← Yêu cầu bổ sung công
├── id, correction_type: attendance|overtime
├── user_id, branch_id
├── attendance_log_id → attendance_logs.id   (nullable, dùng cho ca chính thức)
├── overtime_request_id → overtime_requests.id (nullable, dùng cho tăng ca)
├── original_status, credit_count, description
├── status: pending|approved|rejected
├── processed_by_id, processed_at, manager_note
├── deleted_at
├── UNIQUE(attendance_log_id) WHERE NOT NULL AND deleted_at IS NULL
└── UNIQUE(overtime_request_id) WHERE NOT NULL AND deleted_at IS NULL

leave_requests             ← Yêu cầu nghỉ phép
├── id, user_id, branch_id
├── leave_date, leave_type: full_day|half_day_morning|half_day_afternoon
├── time_from, time_to, original_status, description
├── status: pending|approved|rejected
├── processed_by_id, processed_at, manager_note
├── deleted_at
└── UNIQUE(user_id, leave_date)

overtime_requests          ← Yêu cầu tăng ca
├── id, user_id, branch_id, date
├── actual_checkin, actual_checkout     ← Thời gian thực tế (timestamptz)
├── calculated_start, calculated_end   ← Sau bo tròn (tính khi duyệt)
├── total_hours                        ← decimal(5,2)
├── status: init|pending|approved|rejected
├── manager_id, processed_at, manager_note
├── deleted_at
└── UNIQUE(user_id, date)

daily_summaries            ← Pre-computed aggregate, thay thế query nặng trên attendance_logs
├── id, branch_id → branches.id, date
├── total_employees, present_count, late_count
├── early_leave_count, half_day_count, absent_count
├── total_work_hours, total_overtime, fraud_count
├── attendance_rate (%), on_time_rate (%)
├── computed_at, deleted_at
└── UNIQUE(branch_id, date)
```

### Index Strategy (tối ưu cho 5000 users)

```sql
-- users
idx_user_branch_role_active    (branch_id, role, is_active)       -- RBAC filter chính
idx_user_branch_active         (branch_id, is_active)             -- Manager xem chi nhánh

-- attendance_logs (hot table)
uniq_attendance_user_date      (user_id, date)   UNIQUE           -- Race condition check-in
idx_attendance_user_date       (user_id, date DESC)               -- Lịch sử cá nhân
idx_attendance_branch_date     (branch_id, date DESC)             -- Manager dashboard
idx_attendance_branch_status_date (branch_id, status, date)       -- Báo cáo theo trạng thái
idx_attendance_fraud           (user_id, created_at DESC)
    WHERE is_fake_gps = TRUE OR is_vpn = TRUE                     -- Partial index — chỉ bản ghi nghi ngờ

-- daily_summaries
uniq_daily_branch_date         (branch_id, date) UNIQUE           -- Upsert hàng ngày
idx_daily_branch_date_range    (branch_id, date)                  -- Báo cáo xu hướng 30 ngày
idx_daily_date                 (date)                             -- Admin xem toàn hệ thống theo ngày

-- wifi_configs / gps_configs / shifts
idx_wifi_branch_active         (branch_id, is_active)             -- Hot path check-in
idx_wifi_bssid                 (bssid)  WHERE bssid != ''         -- Lookup BSSID
idx_gps_branch_active          (branch_id, is_active)             -- Hot path khi WiFi fail
idx_shift_branch_default_active (branch_id, is_default, is_active)
```

---

## Bảo mật và Anti-Fraud

### Luồng đầy đủ khi Check-in

```
POST /attendance/check-in
        │
        ▼
[1] JWT Auth                     → 401 nếu token hết hạn / không hợp lệ
        │
        ▼
[2] Rate Limit (10/phút/user)    → 429 nếu vượt ngưỡng
        │
        ▼
[3] Load user từ DB              → Lấy BranchID authoritative (không tin client)
        │
        ▼
[4] Fake GPS (mobile SDK flag)   → 403 FAKE_GPS_DETECTED
        │
        ▼
[5] VPN (mobile SDK flag)        → 403 VPN_DETECTED
        │
        ▼
[6] IP Blocklist (Redis)         → 403 VPN_DETECTED (fail-open nếu Redis lỗi)
        │
        ▼
[7] Suspicious count ≥ 3/7 ngày → 403 TOO_MANY_SUSPICIOUS
        │
        ▼
[8] Đã check-in hôm nay?        → 409 ALREADY_CHECKED_IN
        │
        ▼
[9] Load Shift mặc định          → Xác định khung giờ tính muộn/về sớm
        │
        ▼
[10] Validate WiFi hoặc GPS      → 403 LOCATION_INVALID
        │
        ▼
[11] Tính Status + Work Hours
        │
        ▼
[12] Save DB → SET Cache
        │
        ▼
     200 OK ✓
```

---

## Tối ưu hiệu năng

### Database
- **Connection Pool**: `MaxOpenConns=25`, `MaxIdleConns=10`, `ConnMaxLifetime=300s`
- **Composite Indexes**: Index theo đúng access pattern thực tế, không index thừa
- **Partial Index**: Chỉ index bản ghi gian lận → nhỏ hơn, nhanh hơn
- **CTE SQL**: Dashboard today dùng Common Table Expression tính song song — một câu SQL, không N+1
- **UNIQUE constraint**: `(user_id, date)` ngăn race condition khi 2 request check-in đồng thời
- **DailySummary table**: Pre-computed aggregate tránh scan hàng triệu rows mỗi lần load dashboard

### Redis Caching
- **Cache-aside + SET sau write**: Sau check-in/check-out, SET cache ngay giá trị mới thay vì chỉ DELETE → request tiếp theo hit cache
- **TTL phân tầng**: Data thay đổi nhiều (2 phút) vs ít thay đổi (15 phút)

### Application
- **Fail-open Rate Limiter**: Redis lỗi → cho request qua, không block toàn service
- **Async Last Login**: Goroutine fire-and-forget, không block login response
- **Graceful Shutdown**: Context timeout 10s — chờ request đang xử lý hoàn thành

---

## Tech Stack

| Thành phần | Công nghệ | Phiên bản |
|------------|-----------|-----------|
| Language | Go | 1.22 |
| Web framework | Echo | v4 |
| ORM | GORM | v2 |
| Database | PostgreSQL | 16 |
| Cache | Redis | 7 |
| Migration | gormigrate | v2 |
| Auth | golang-jwt | v5 |
| Password | bcrypt | cost 10 |
| Logging | slog | stdlib |
| Config | Viper | — |
| Container | Docker + Compose | — |
