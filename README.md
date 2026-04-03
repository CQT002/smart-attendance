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
8. [Tính năng & Lộ trình phát triển](#8-tính-năng--lộ-trình-phát-triển)
9. [Cài đặt & Chạy Project](#9-cài-đặt--chạy-project)

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
- **Phát hiện GPS giả:** Kiểm tra cờ tệp (flag) `is_fake_gps` từ Mobile SDK.
- **Phát hiện VPN/Proxy:** Khớp cờ `is_vpn` để chặn điểm danh.
- **IP Blocklist:** Khớp IP yêu cầu với danh sách đen đã khai báo.
- **Lịch sử vi phạm (Behavior):** Chặn nếu tài khoản vi phạm trên 3 lần trong 7 ngày.

> **Thiết kế quan trọng:** Dù phát hiện gian lận bằng SDK, hệ thống vẫn cho phép tạo bản ghi nhưng lưu trạng thái "nghi ngờ" để Manager rà soát thay vì tự động khóa ngay, tránh các trường hợp block nhầm.

#### Lớp 2 — Location Validation
- **Xác thực WiFi (Primary):** Khớp SSID và BSSID gửi lên với cấu hình thiết bị phát mạng lưu trữ tại cơ sở dữ liệu chi nhánh.
- **GPS Geofencing (Fallback):** Sử dụng hệ thức khoảng cách điểm (Haversine Formula) đối chiếu toạ độ thiết bị trong bán kính quy định (thường là ±10m).

#### Lớp 3 — Business Rules & Security
- **Ngăn chặn leo thang đặc quyền (Privilege Escalation):** Toàn bộ dữ liệu chi nhánh được đối chiếu qua JWT Token và truy vấn từ Database, nghiêm cấm đặt lòng tin vào tham số `branch_id` do thiết bị tự truyền xuống.
- **Khóa Spam/Rate Limit:** Cập nhật phiên bằng Database và cấu trúc bộ nhớ đệm (Cache) để chống gửi liên tục.

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

Kiến trúc cách ly dữ liệu Database được thiết kế cho sự tương tác giữa hàng chục ngàn nhân viên đồng thời:
- **Cấu trúc Ràng buộc (Schema constraint):** Dữ liệu User bị khóa cứng với Role tương ứng và Branch. Các thao tác đều phụ thuộc khóa ngoại.
- **Chỉ mục tối ưu (Partial Indexing):** Hệ thống chỉ tạo Index chuyên biệt trên các bản ghi có sự mâu thuẫn (như Fake GPS/VPN). Kỹ thuật này giữ cho tốc độ nhận diện sự cố đạt tốc độ chớp nhoáng (O(log M)).

### Báo cáo Tốc độ cao (Pre-computed Aggregates)

Thay vì gọi hàm toán học Group-By tốn kém tài nguyên mỗi khi yêu cầu báo cáo, cấu trúc bảng `daily_summaries` lưu trữ các dữ liệu tích lũy chạy theo cơ chế "Cập nhật bù" (UPSERT). Hệ thống tự cộng dồn chỉ số khi có điểm danh mới.

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

**Quy trình thực thi ngăn lách luật:**
- **Tầng Middleware:** Bốc tách Branch Identifier từ JWT.
- **Tầng Handler Business:** Xác thực vai trò. Nếu Manager yêu cầu tương tác dữ liệu vượt khỏi quyền quản lý nhóm, huỷ ngay giao dịch.
- **Tầng Repository:** Tự động đính kèm mệnh đề điều kiện rào chắn `WHERE branch_id = XYZ` vào đuôi của tiến trình truy cập nhằm không cho hacker nào có cơ hội qua mặt.

### Dashboard Thời gian thực (Loại bỏ N+1 Query)

Dashboard toàn cục dành cho quyền Admin sử dụng bộ công thức Common Table Expressions (CTE) kết xuất chung tỷ lệ gian lận, số người đi làm của cả nền tảng đa chi nhánh chỉ bằng một bước lọc, loại bỏ thao tác vòng lặp đệ quy N+1 Request gây quá tải CPU Database.

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

Token payload (Mã thông báo JWT) không thể bị làm giả vì nó đã được bọc khóa mã hóa liên kết (Signed Secret). Thông tin Phân quyền duy trì trực tiếp trong đó mà không chịu sự can thiệp từ Client đầu cuối. Các API luôn bóc tách thông qua khóa để cấp phép phân lớp dữ liệu.

Token payload **không thể bị giả mạo** (signed HS256). Handler đọc `BranchID` từ token, không từ request body/query.

### Middleware RBAC — Thực thi tại Router Layer

Các tính năng quan trọng và nhạy cảm (Tạo/Xóa User, Cập nhật cấu hình) sẽ đi kèm bộ màng lọc (Middleware) định hướng tự động nhận diện cấp đặc quyền. Bất kỳ sự dồn ép truy cập cấm nào sẽ bị Router Backend triệt tiêu tức thì.

### RBAC trên Frontend

Giao diện App và Web đồng bộ rà soát Rule truy cập từ Token nội bộ, tự ẩn các nút chức năng ngoại tuyến đối với Manager quản lý cấp dưới. Việc bảo vệ Layout Root ngăn việc phơi bày giao diện nhạy cảm dù người dùng chưa có cơ hội tải trang dữ liệu.

---

## 5. Chiến lược Scaling — 5.000 nhân viên đồng thời

### Bài toán Peak Load

> **Kịch bản khắc nghiệt nhất:** 5.000 nhân viên check-in đồng loạt lúc 8:00 SA.
> Mỗi check-in kích hoạt: 1 Rate Limit check + 1 Fraud check + 1 WiFi/GPS query + 1 INSERT.

**Phân tích:**
- ~500 requests/phút trong 10 phút đầu giờ = ~8.3 req/giây
- Mỗi request cần ~3–5 DB queries + ~2 Redis ops

### Bộ nhớ đệm Redis
Hàng rào kỹ thuật phòng bị sự cố (Connection Pool / Pooling) liên tục cho phép duy trì và tái chế luồng liên kết nhàn rỗi trong khoảng thời gian chờ (Timeout limit) nhằm chặn đứng triệt để nguy cơ trút tràn bộ nhớ.

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

#### Trượt băng thông tốc độ cao (Sliding Window Limits)

Cơ chế trượt thanh kết hợp thuật toán bộ nhớ nguyên tử (Atomic Memory Increment) phân tách riêng mức giới hạn nghẽn cục bộ: Điểm danh sở hữu khóa tạm thời là 1 phút/1 lần; Đăng nhập được khóa theo 15 phút. Toàn hệ thống được rào chắn bởi màng bọc phân giải truy cập Global Rate Limit.

### PostgreSQL — Kết nối đa chiều

Kỹ thuật MaxOpenConns linh động duy trì được sức cản chịu tải nặng. Kỹ thuật MaxIdleConns ngăn không cho Database dập tắt vội kết nối (tái sử dụng lượng Idle Pooling dự trữ), giúp Database đáp ứng phản xạ nhanh ở mili-giây.

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

Kỹ thuật **Partial Index** chỉ được lót dữ liệu dành cho log danh sách đánh cờ vi phạm (chiếm <1% lượng dữ liệu toàn bộ). Kỹ thuật O(log N+1) loại hình siêu nhẹ này bảo trì việc lưu trữ thông suốt cho Data Warehouse.

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

### Hệ sinh thái Môi trường (Dockerization & Containers)

Hệ thống được đóng gói đa môi trường (Containerized) bằng quy Standard cấu trúc file tự động thay vì dàn trải hệ điều hành gốc (Native Install). Backend sử dụng màng chắn lớp Multi-Stage gộp cấu thành (Binary Compiling) để ép xung và giảm nhẹ dung lượng tải vận hành Runtime xuống không tưởng (15MB), đảm bảo cho luồng CI/CD lên Server chỉ kéo dài từ 5-10 giây khởi tạo.

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

## 8. Tính năng & Lộ trình phát triển

Hệ thống hiện tại đã đáp ứng Core Flow xuất sắc cho việc điểm danh bảo mật. Dưới đây là danh sách tính năng hiện hữu và lộ trình nâng cấp thành Hệ thống Quản trị Nhân sự (HRIS) toàn diện.

### 8.1. Các tính năng hiện hữu (Current Features)

#### Employee App (Mobile)
- **Điểm danh an toàn:** Hỗ trợ check-in/out qua GPS và WiFi (SSID/BSSID).
- **Anti-Fraud:** Chống giả mạo vị trí qua cơ chế nhận diện Fake GPS và Fake VPN.
- **Lịch sử cá nhân:** Xem trạng thái điểm danh hôm nay và lịch sử theo thời gian.
- **Bảo mật thiết bị:** Xác định danh tính qua Device ID, App Version.

#### Admin Portal (Web)
- **Real-time Dashboard:** Thống kê tỷ lệ đi làm, đi muộn, vắng mặt.
- **Quản lý đa cấu hình:** Quản lý Chi nhánh, thiết lập mạng WiFi hợp lệ, cấu hình toạ độ Geofencing.
- **Quản lý Nhân sự & Phân quyền:** RBAC (Admin/Manager), tạo/sửa/xoá nhân viên, cấp lại mật khẩu.
- **Traceability:** Xem lịch sử mọi bản ghi điểm danh và dấu hiệu bị nghi ngờ gian lận.

### 8.2. Lộ trình phát triển (HRIS Roadmap)

#### Nhóm Dành cho Nhân viên (Employee)
- **Log chấm công bù (Manual Log):** Đơn giải trình quên check-in/out hoặc đi muộn vì lý do khách quan (hỏng xe, gặp đối tác).
- **Chấm công theo ca (Shift Scheduling):** Hỗ trợ linh hoạt định nghĩa các loại ca (ca hành chính, ca gãy, ca đêm, xoay ca).
- **Quản lý Nghỉ phép (Leave Management):** Gửi và theo dõi tình trạng đơn xin nghỉ (phép năm, thai sản, nghỉ ốm).
- **Đăng ký OT (Overtime):** Khai báo làm thêm giờ đi kèm theo hệ số lương và lý do.
- **Đi công tác (Business Trip):** Đăng ký lịch trình và check-in tại vị trí công tác ngoại tuyến.

#### Nhóm Dành cho Quản lý (Admin/Manager)
- **Phê duyệt trực tuyến (Online Approvals):** Cấp quản lý duyệt hoặc từ chối ngay lập tức các đơn giải trình trực tiếp trên giao diện app hoặc Web.
- **Dashboard Thời gian thực nâng cao:** Bố trí danh sách "live" chi tiết đi muộn, vắng mặt dành riêng cho Manager chi nhánh trên App
- **Quản lý thiết bị (Device Whitelist):** Gỡ liên kết / Phê duyệt máy mới khi nhân viên đổi điện thoại để tránh việc điểm danh hộ.
- **Xuất báo cáo tự động (Export Report):** Tự động kết xuất báo cáo tổng hợp ra file Excel chuyên nghiệp vào mỗi cuối tháng.

#### Nhóm Nâng cao Chống Gian Lận (Anti-Fraud & AI)
- **Face Liveness / Recognition:** Yêu cầu chụp ảnh Selfie khi check-in, dùng AI so khớp với khuôn mặt thật để chặn triệt để tình trạng người khác bấm hộ trên máy dự phòng.
- **Thuật toán rà soát Di chuyển ngầm:** Chạy CronJob phân tích "Điểm chốt Tọa độ & Cảm biến bước chân offline". Nếu một thiết bị liên tục báo về nằm im tại khu vực văn phòng 3 ngày liền mạch (độ chênh lệch vận động = 0), lập tức tống xuất vào Danh sách Đỏ (Máy nghi ngờ gian lận) gửi thông báo cho Manager kiểm tra đột xuất tại chi nhánh.

#### Nhóm Khác (Extensions)
- **Bảo mật Phiếu Lương (Secure Payslip):** Xem chi tiết bảng lương ngay trên App yêu cầu trích xuất bằng mã PIN / FaceID/TouchID an toàn.
- **Email Tự động cảnh báo (Alert Automation):** Tự định tuyến gửi email danh sách nhân viên vi phạm / nghi ngờ sử dụng Mock GPS cho trưởng chi nhánh mỗi sáng/tuần.

---

## 9. Cài đặt & Chạy Project

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
