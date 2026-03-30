# PROMPT_LOG.md — HDBBank Smart Attendance

> Ghi lại các prompt đã sử dụng để xây dựng dự án từ đầu với sự hỗ trợ của AI.
> Mỗi prompt tương ứng với một giai đoạn phát triển cụ thể.

---

## Mục lục

1. [Phase 1: Khởi tạo dự án & Kiến trúc](#phase-1-khởi-tạo-dự-án--kiến-trúc)
2. [Phase 2: Backend API (sa-api)](#phase-2-backend-api-sa-api)
3. [Phase 3: Admin Portal (sa-web)](#phase-3-admin-portal-sa-web)
4. [Phase 4: Mobile App (sa-mb)](#phase-4-mobile-app-sa-mb)
5. [Phase 5: Tài liệu & Hoàn thiện](#phase-5-tài-liệu--hoàn-thiện)

---

## Phase 1: Khởi tạo dự án & Kiến trúc

### Prompt 1.1 — Tạo CLAUDE.md (Quy tắc dự án)

```
Tạo file CLAUDE.md cho dự án "HDBBank Smart Attendance" — hệ thống chấm công thông minh
cho 100 chi nhánh, 5.000 nhân viên.

Tech stack:
- Backend: Go 1.22, Echo v4, GORM v2, PostgreSQL 16, Redis 7, JWT HS256
- Config: Viper + config.yaml
- Logging: log/slog (stdlib)
- Password: bcrypt cost 10

Yêu cầu CLAUDE.md phải bao gồm:
1. Tech stack — bảng liệt kê rõ ràng, không được thêm dependency mới tùy ý
2. Cấu trúc thư mục Clean Architecture với quy tắc dependency giữa các layer
3. Quy trình xử lý lỗi tập trung (AppError, response chuẩn hóa)
4. Logging conventions với slog (level, structured logging)
5. Conventional Commits format (type, scope, ví dụ)
6. Code conventions cho Go, GORM, API routes
7. Security rules (không log sensitive data, không commit secrets)
```

---

## Phase 2: Backend API (sa-api)

### Prompt 2.1 — Khởi tạo project Go & Infrastructure

```
Khởi tạo project Go cho sa-api theo cấu trúc Clean Architecture đã định nghĩa trong CLAUDE.md.

Tạo:
1. go.mod với module name và tất cả dependencies trong tech stack
2. cmd/server/main.go — entrypoint với dependency injection
3. config/config.go — Viper loader đọc config.yaml
4. config/config.yaml — giá trị dev-safe, secret dùng placeholder "change-me-in-production"
5. internal/infrastructure/database/postgres.go — GORM setup với connection pool
   (max_open_conns: 25, max_idle_conns: 10)
6. internal/infrastructure/cache/redis.go — Redis client + Cache interface
7. internal/infrastructure/logger/logger.go — slog setup (JSON cho production, Text cho dev)

Config.yaml cần có sections: app, database, redis, jwt
```

### Prompt 2.2 — Domain Entities (GORM Models)

```
Tạo các GORM entity models trong internal/domain/entity/:

1. user.go — User model với:
   - ID, EmployeeCode, FullName, Email, PasswordHash, Phone
   - Role (admin/manager/employee) — RBAC 3 tầng
   - BranchID (FK), IsActive, LastLoginAt
   - Indexes: uniq_user_email, idx_user_branch_role_active

2. branch.go — Branch model với:
   - ID, Code, Name, Address, IsActive
   - GPS config: Latitude, Longitude, GeofenceRadius (default 100m)
   - WiFi config: có quan hệ 1-nhiều với WifiConfig

3. wifi_config.go — WifiConfig model: BranchID, SSID, BSSID

4. gps_config.go — GPSConfig model: BranchID, Latitude, Longitude, RadiusMeters

5. attendance.go — AttendanceLog model với:
   - UserID, BranchID, Date, CheckInTime, CheckOutTime
   - CheckInMethod/CheckOutMethod (wifi/gps)
   - Status (present/late/absent/half_day)
   - WiFi fields: SSID, BSSID
   - GPS fields: Latitude, Longitude, DistanceMeters
   - Anti-fraud: IsFakeGPS, IsVPN, DeviceID, DeviceModel
   - FraudFlags, Notes, WorkHours
   - Unique constraint: (user_id, date)
   - Indexes cho branch_date, branch_status_date, fraud partial index

6. daily_summary.go — DailySummary: BranchID, Date, TotalEmployees, PresentCount,
   LateCount, AbsentCount, AvgWorkHours

7. shift.go — Shift model: Name, StartTime, EndTime, LateThreshold, HalfDayThreshold

Tất cả dùng gorm.Model (ID, CreatedAt, UpdatedAt, DeletedAt).
Tuân theo convention snake_case, không dùng gorm:"column:..." trừ khi cần.
```

### Prompt 2.3 — Repository Interfaces & Implementations

```
Tạo repository layer:

1. Domain interfaces (internal/domain/repository/):
   - attendance.go: AttendanceRepository interface — Create, Update, FindByID,
     FindByUserAndDate, GetList (with filters, pagination), GetTodayStatsByBranch,
     CountByDateRange
   - user.go: UserRepository interface — Create, Update, Delete, FindByID, FindByEmail,
     GetList (with filters, pagination), CountByBranch
   - branch.go: BranchRepository interface — Create, Update, Delete, FindByID,
     GetList, GetActiveBranches
   - wifi_config.go, gps_config.go, shift.go interfaces

2. PostgreSQL implementations (internal/repository/):
   - Implement tất cả interfaces trên
   - Luôn dùng .WithContext(ctx) trước mọi query
   - Dùng Preload cho quan hệ 1-nhiều
   - Dùng raw SQL cho aggregate queries phức tạp
   - Trả về apperrors.Wrap() khi có lỗi DB
   - helper.go cho các database helper functions
```

### Prompt 2.4 — AppError & Response Package

```
Tạo pkg/ packages:

1. pkg/apperrors/errors.go:
   - AppError struct: HTTPStatus, Code, Message, Err (wrapped)
   - func New(status, code, message) *AppError
   - func Wrap(err, status, code, message) *AppError
   - Implement error interface và Unwrap()
   - Predefined errors: ErrUserNotFound, ErrInvalidCredentials, ErrUnauthorized,
     ErrForbidden, ErrDuplicateEmail, ErrBranchNotFound, ErrAlreadyCheckedIn,
     ErrFakeGPSDetected, ErrVPNDetected, ErrOutOfGeofence, ErrWiFiNotMatched,
     ErrRateLimited, ErrAccountLocked
   - ValidationError cho input validation

2. pkg/response/response.go:
   - OK(c, data), Created(c, data), NoContent(c)
   - OKWithMessage(c, message, data)
   - Paginated(c, items, total, page, limit)
   - Error(c, err) — tự detect AppError vs generic error

3. pkg/utils/:
   - jwt.go: GenerateToken, ParseToken dùng golang-jwt/jwt/v5
   - geo.go: Haversine formula tính khoảng cách GPS, IsWithinGeofence
   - pagination.go: ParsePagination từ query params, PaginationParams struct
```

### Prompt 2.5 — Middleware Stack

```
Tạo internal/middleware/:

1. auth.go — JWT Authentication middleware:
   - JWTAuth(): Extract Bearer token từ Authorization header, parse và validate
   - Set user_id, email, role vào echo.Context
   - RequireRole(roles ...string): RBAC middleware — check role có trong danh sách cho phép
   - Trả 401 nếu token invalid, 403 nếu role không đủ quyền

2. rate_limiter.go — Redis-based rate limiting:
   - LoginRateLimiter: 10 requests / 15 phút per IP (sliding window)
   - CheckInRateLimiter: 10 requests / 1 phút per user_id
   - Dùng Redis INCR + EXPIRE cho atomic counter

3. request_logger.go — Structured request logging:
   - Generate request_id (UUID) cho mỗi request
   - Log method, path, status, latency
   - slog.Info cho 2xx-3xx, slog.Warn cho 4xx, slog.Error cho 5xx
   - Set request_id vào response header X-Request-ID
```

### Prompt 2.6 — Usecase: User (Auth & Management)

```
Tạo internal/usecase/user/main.go:

1. Domain interfaces (internal/domain/usecase/user.go):
   - LoginRequest: Email, Password
   - LoginResponse: AccessToken, RefreshToken, User
   - CreateUserRequest: EmployeeCode, FullName, Email, Password, Phone, Role, BranchID
   - UpdateUserRequest: FullName, Phone, Role, BranchID, IsActive
   - UserUsecase interface: Login, GetMe, ChangePassword, CreateUser, UpdateUser,
     DeleteUser, GetUserByID, GetUsers (paginated), ResetPassword

2. Implementation:
   - Login: FindByEmail → bcrypt.CompareHashAndPassword → GenerateToken (24h access, 7d refresh)
     → goroutine UpdateLastLogin (fire-and-forget)
   - CreateUser: hash password bcrypt cost 10, check duplicate email
   - Admin login riêng: 403 nếu role = employee
   - ChangePassword: verify old password trước khi update
   - ResetPassword: admin only, set password mới
```

### Prompt 2.7 — Usecase: Attendance (Check-in/out với Anti-fraud)

```
Tạo internal/usecase/attendance/main.go:

Domain interfaces (internal/domain/usecase/attendance.go):
- CheckInRequest: WiFi (SSID, BSSID), GPS (Latitude, Longitude),
  DeviceID, DeviceModel, IsFakeGPS, IsVPN
- CheckOutRequest: tương tự CheckInRequest
- AttendanceUsecase interface: CheckIn, CheckOut, GetMyToday, GetList, GetByID, GetSummary

Luồng Check-in (3 lớp validation tuần tự):

Layer 1 — Anti-fraud:
  - Check IsFakeGPS → reject nếu true
  - Check IsVPN → reject nếu true
  - Check DeviceID suspicious count trong 7 ngày gần nhất
  - Nếu ≥ 3 fraud flags → temp lock account

Layer 2 — Location Validation:
  - Lấy branch config (WiFi list + GPS config) từ user's BranchID
  - WiFi validation: match SSID + BSSID với danh sách WiFi của branch
  - GPS validation: Haversine distance ≤ GeofenceRadius
  - Ưu tiên WiFi (chính xác hơn indoor), fallback GPS

Layer 3 — Business Rules:
  - Check đã check-in hôm nay chưa (unique user_id + date)
  - Xác định status dựa trên shift config:
    - CheckIn ≤ StartTime + LateThreshold → present
    - CheckIn > LateThreshold → late
  - Tạo AttendanceLog record

Check-out:
  - Tìm record check-in hôm nay
  - Cập nhật CheckOutTime, tính WorkHours
  - Nếu WorkHours < HalfDayThreshold → half_day
```

### Prompt 2.8 — Usecase: Branch & Report

```
Tạo 2 usecase:

1. internal/usecase/branch/main.go:
   - Domain interfaces (internal/domain/usecase/branch.go):
     CreateBranchRequest, UpdateBranchRequest, BranchFilter
   - CRUD operations với Redis cache cho GetActiveBranches
   - Cache invalidation khi create/update/delete

2. internal/usecase/report/main.go:
   - Domain interfaces (internal/domain/usecase/report.go):
     DashboardStats, TodayBranchStats, ReportFilter, EmployeeReport, BranchReport
   - DashboardStats: 6 KPI metrics (total employees, present today, late, absent,
     on-time rate, avg work hours)
   - TodayBranchStats: thống kê theo branch (dùng CTE aggregate query)
   - Attendance report: filter theo date range, branch, user — hỗ trợ group by employee
     hoặc branch
   - Individual user report: lịch sử chấm công của 1 user
```

### Prompt 2.9 — HTTP Handlers

```
Tạo HTTP handlers:

1. internal/handler/admin/:
   - auth.go: Login (403 cho employee), Me, ChangePassword
   - user.go: CRUD employees — GetList (paginated, filter by branch/role/search),
     Create, GetByID, Update, Delete (admin only), ResetPassword
   - branch.go: CRUD branches — GetList, GetActiveBranches, Create, GetByID,
     Update, Delete (admin only)
   - attendance.go: GetList (filter by date/branch/status/user), GetByID, GetSummary
   - report.go: GetDashboardStats, GetTodayStats, GetTodayEmployees,
     GetAttendanceReport, GetBranchReport, GetUserReport

2. internal/handler/user/:
   - auth.go: Login, Me, ChangePassword
   - attendance.go: CheckIn, CheckOut, GetMyToday
   - helper.go: getUserIDFromContext()

Handler rules:
- Parse request → call usecase → dùng response.OK/Error
- Không tự format JSON, luôn dùng pkg/response
- Input validation tại handler level
- Manager chỉ thấy data của branch mình (filter BranchID từ context)
```

### Prompt 2.10 — Router & Server Setup

```
Tạo internal/server/router.go:

Route registration cho Echo:
- Health check: GET /health
- Employee routes: /api/v1/auth/*, /api/v1/attendance/*
- Admin routes: /api/v1/admin/* (tất cả require JWT)
- Middleware stack: CORS, Recover, RequestLogger
- Route groups với middleware:
  - Login routes: + LoginRateLimiter
  - Auth routes: + JWTAuth
  - Admin routes: + JWTAuth + RequireRole("admin", "manager")
  - Admin-only routes: + RequireRole("admin")
  - Check-in routes: + CheckInRateLimiter

Cập nhật cmd/server/main.go:
- Load config → Init logger → Connect DB → Connect Redis → AutoMigrate
- Inject dependencies: repos → usecases → handlers
- Register routes → Start server với graceful shutdown (10s timeout)
```

### Prompt 2.11 — Database Migrations

```
Tạo hệ thống migration:

1. internal/infrastructure/database/migrations/migrations.go:
   - Dùng gormigrate v2
   - Migration 001: AutoMigrate tất cả entities
   - Tạo indexes đã định nghĩa trong entity

2. cmd/migration/main.go:
   - CLI tool cho manual migration (up/down/status)

3. migrations/001_init_schema.sql:
   - SQL reference file cho documentation

4. docker-compose.yml:
   - PostgreSQL 16 service (port 5432)
   - Redis 7 service (port 6379)
   - Volume mounts cho data persistence

5. Dockerfile:
   - Multi-stage build: builder (Go compile) → runtime (scratch/alpine)

6. Makefile:
   - Targets: build, run, migrate, docker-up, docker-down, test
```

---

## Phase 3: Admin Portal (sa-web)

### Prompt 3.1 — Khởi tạo Next.js Project

```
Khởi tạo sa-web — Admin Portal cho HDBBank Smart Attendance:

Tech stack:
- Next.js 15 (App Router, Turbopack)
- TypeScript 5
- Tailwind CSS 3.4
- shadcn/ui (Radix UI primitives)
- TanStack Query v5 (server-state management)
- Axios (HTTP client với JWT interceptor)
- react-hook-form + Zod (forms & validation)
- Recharts (charts)
- date-fns (date formatting, Vietnamese locale)
- Sonner (toast notifications)
- js-cookie (JWT cookie management)
- lucide-react (icons)

Tạo:
1. Project structure với App Router
2. tailwind.config.ts với CSS variables cho theming (light/dark)
3. globals.css với shimmer animation keyframe
4. tsconfig.json với strict mode, @/* path alias
5. components.json cho shadcn/ui config
```

### Prompt 3.2 — TypeScript Types & API Client

```
Tạo types và API infrastructure:

1. types/ — TypeScript interfaces matching backend:
   - api.ts: ApiResponse<T>, PaginatedResponse<T>, PaginationMeta
   - auth.ts: UserRole enum, User, LoginRequest, LoginResponse
   - user.ts: User, CreateUserRequest, UpdateUserRequest, UserFilter
   - branch.ts: Branch, WifiNetwork, GPSConfig, CreateBranchRequest, BranchFilter
   - attendance.ts: AttendanceLog, AttendanceStatus enum, AttendanceFilter
   - report.ts: DashboardStats, TodayBranchStats, ReportFilter, EmployeeReport
   - index.ts: re-export all

2. lib/api-client.ts:
   - Axios instance với baseURL từ NEXT_PUBLIC_API_URL
   - Request interceptor: auto-add Bearer token từ cookie
   - Response interceptor: 401 → clear cookies → redirect /login
   - Generic request helpers: get<T>, post<T>, put<T>, delete<T>

3. lib/auth.ts:
   - isAuthenticated(): check cookie exists
   - getStoredUser(): từ localStorage
   - clearStoredUser(): remove cookies + localStorage

4. lib/utils.ts:
   - cn(): clsx + tailwind-merge
   - formatDate(), formatDateTime(): date-fns với Vietnamese locale
   - formatPercent(): number formatting
```

### Prompt 3.3 — API Services & TanStack Query Hooks

```
Tạo service layer và hooks:

1. services/:
   - auth.service.ts: login (save tokens to cookies + user to localStorage),
     getMe, changePassword, logout
   - user.service.ts: getList (paginated, filtered), create, update, delete,
     getById, resetPassword
   - branch.service.ts: getList, getActive (cho dropdowns), create, update, delete, getById
   - attendance.service.ts: getList (filtered), getById, getSummary
   - report.service.ts: getDashboardStats, getTodayStats, getTodayEmployees,
     getAttendanceReport, getBranchReport, getUserReport

2. hooks/ — TanStack Query wrappers:
   - use-auth.ts: useCurrentUser, useLogin (mutation), useLogout (mutation),
     useChangePassword
   - use-users.ts: useUsers (query, paginated), useCreateUser, useUpdateUser,
     useDeleteUser, useResetPassword — tất cả invalidate query on success
   - use-branches.ts: useBranches, useActiveBranches (staleTime: 5min),
     useCreateBranch, useUpdateBranch, useDeleteBranch
   - use-attendance.ts: useAttendanceLogs, useAttendanceDetail, useAttendanceSummary
   - use-reports.ts: useDashboardStats (refetchInterval: 5min),
     useTodayBranchStats, useTodayEmployees, useAttendanceReport, useBranchReport

Mỗi mutation hook: onSuccess → toast success + invalidateQueries, onError → toast error
```

### Prompt 3.4 — UI Components (shadcn/ui)

```
Tạo shadcn/ui components trong components/ui/:

1. button.tsx — variants: default, destructive, outline, secondary, ghost, link
   + sizes: default, sm, lg, icon
2. card.tsx — Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter
3. input.tsx — styled input với focus ring
4. label.tsx — Radix Label
5. select.tsx — Radix Select với trigger, content, item, group
6. dialog.tsx — Radix Dialog với overlay, content, header, footer, title, description
7. table.tsx — Table, TableHeader, TableBody, TableRow, TableHead, TableCell
8. badge.tsx — variants: default, secondary, destructive, outline,
   success (green), warning (yellow/orange), info (blue)
9. skeleton.tsx — shimmer animation dùng CSS keyframe từ globals.css

Tất cả dùng React.forwardRef, className merge bằng cn().
```

### Prompt 3.5 — Layout Components

```
Tạo layout components:

1. components/layout/sidebar.tsx:
   - Collapsible sidebar (expand/collapse toggle)
   - Navigation links: Dashboard, Chi nhánh, Nhân viên, Chấm công, Báo cáo
   - Active link highlight dựa trên pathname
   - Responsive: overlay trên mobile, persistent trên desktop
   - User info + logout button ở bottom
   - Icons từ lucide-react

2. components/layout/header.tsx:
   - Page title (dynamic theo route)
   - Current user info (name, role badge)
   - Mobile menu toggle button

3. app/layout.tsx:
   - Root layout: Inter font, metadata
   - Wrap children với Providers

4. app/providers.tsx:
   - QueryClientProvider (TanStack Query)
   - Sonner Toaster config
   - ThemeProvider (nếu cần)

5. app/(auth)/layout.tsx — layout không sidebar cho login
6. app/(admin)/layout.tsx — layout có sidebar, check isAuthenticated → redirect /login
```

### Prompt 3.6 — Shared Components

```
Tạo components/shared/:

1. pagination.tsx:
   - Smart pagination: First, Prev, page numbers, Next, Last
   - Hiển thị range: "Hiển thị 1-20 của 5.000 kết quả"
   - Props: page, limit, total, onPageChange

2. status-badge.tsx:
   - StatusBadge: map AttendanceStatus → badge variant + label tiếng Việt
     (present→green "Đúng giờ", late→orange "Trễ", absent→red "Vắng",
     half_day→blue "Nửa ngày")
   - RoleBadge: admin→destructive, manager→warning, employee→default
   - ActiveBadge: active→green, inactive→gray

3. data-table-skeleton.tsx:
   - Skeleton loading cho table: header + N rows với shimmer effect
   - Props: columns count, rows count
```

### Prompt 3.7 — Trang Login

```
Tạo app/(auth)/login/page.tsx:

- Form đăng nhập với email + password
- Validation bằng Zod: email format required, password ≥ 6 ký tự
- react-hook-form integration
- Submit → gọi useLogin mutation
- Error display từ API (toast)
- Loading state trên button
- Redirect /dashboard sau login thành công
- Layout: centered card, logo HDBBank (hoặc tên hệ thống), background gradient
```

### Prompt 3.8 — Trang Dashboard

```
Tạo app/(admin)/dashboard/page.tsx:

Dashboard hiển thị tổng quan real-time:

1. 6 KPI Cards (2 rows x 3 cols):
   - Tổng nhân viên (icon Users)
   - Có mặt hôm nay (icon UserCheck, green)
   - Đi trễ (icon Clock, orange)
   - Vắng mặt (icon UserX, red)
   - Tỷ lệ đúng giờ (icon TrendingUp, %)
   - Giờ làm trung bình (icon Timer)
   - Skeleton loading cho mỗi card

2. Charts section (2 cols):
   - Pie chart: phân bố trạng thái (present/late/absent) — Recharts PieChart
   - Bar chart: top 8 chi nhánh theo tỷ lệ chấm công — Recharts BarChart

3. Table: Chi nhánh hôm nay
   - Columns: Tên chi nhánh, Tổng NV, Có mặt, Trễ, Vắng, Tỷ lệ
   - Auto-refresh mỗi 5 phút (refetchInterval: 300000)
   - Skeleton khi loading

Data từ hooks: useDashboardStats, useTodayBranchStats
```

### Prompt 3.9 — Trang Quản lý Chi nhánh

```
Tạo app/(admin)/branches/page.tsx + components/branches/branch-form-dialog.tsx:

1. Page:
   - Search bar (tìm theo tên/mã chi nhánh)
   - Button "Thêm chi nhánh" (admin only)
   - Table: Mã, Tên, Địa chỉ, GPS (lat/lng), Bán kính, Trạng thái, Actions
   - Actions: Edit, Delete (admin only) — confirm dialog trước khi xóa
   - Pagination component
   - Skeleton loading

2. BranchFormDialog:
   - Mode: Create / Edit (prefill data)
   - Fields: Code, Name, Address, Latitude, Longitude, GeofenceRadius
   - Zod validation
   - Submit → useCreateBranch hoặc useUpdateBranch
   - Toast success/error
   - Close dialog + refetch list on success
```

### Prompt 3.10 — Trang Quản lý Nhân viên

```
Tạo app/(admin)/users/page.tsx + components/users/:

1. Page:
   - Filter bar: search (tên/email), branch dropdown (useActiveBranches), role filter
   - Button "Thêm nhân viên" (admin only)
   - Table: Mã NV, Họ tên, Email, Chi nhánh, Vai trò, Trạng thái, Actions
   - Actions: Edit, Reset Password, Delete (admin only)
   - Pagination cho 5.000+ records
   - Skeleton loading

2. components/users/user-form-dialog.tsx:
   - Mode: Create / Edit
   - Fields: EmployeeCode, FullName, Email, Phone, Password (create only),
     Role (select), BranchID (select from active branches)
   - Zod validation (email format, password ≥ 6)
   - Submit → useCreateUser / useUpdateUser

3. components/users/reset-password-dialog.tsx:
   - Input new password
   - Confirm action
   - Submit → useResetPassword
```

### Prompt 3.11 — Trang Chấm công

```
Tạo app/(admin)/attendance/page.tsx:

- Filter bar:
  - Date range picker (từ ngày — đến ngày)
  - Branch dropdown
  - Status filter (all/present/late/absent/half_day)
  - Search (tên nhân viên)

- DataTable:
  - Columns: Nhân viên, Chi nhánh, Ngày, Giờ vào, Giờ ra, Phương thức,
    Trạng thái, Giờ làm, Fraud flags
  - StatusBadge cho trạng thái
  - Hiển thị method (WiFi/GPS) với icon
  - Fraud flags: icon cảnh báo nếu có IsFakeGPS hoặc IsVPN

- Pagination
- Skeleton loading
- Click row → xem chi tiết (optional dialog)
```

### Prompt 3.12 — Trang Báo cáo

```
Tạo app/(admin)/reports/page.tsx:

1. Filter section:
   - Period: Hôm nay / Tuần này / Tháng này / Tùy chọn
   - Date range picker (cho tùy chọn)
   - Branch filter
   - Tabs: Theo nhân viên / Theo chi nhánh

2. Tab "Theo nhân viên":
   - Table: Nhân viên, Chi nhánh, Số ngày làm, Đúng giờ, Trễ, Vắng,
     Tỷ lệ chuyên cần, Giờ TB
   - Color coding: >90% green, 70-89% yellow, <70% red
   - Bar chart: top/bottom performers

3. Tab "Theo chi nhánh":
   - Table: Chi nhánh, Tổng NV, Có mặt TB, Trễ TB, Vắng TB, Tỷ lệ
   - Bar chart: so sánh chi nhánh

4. KPI summary cards ở top

Data từ: useAttendanceReport, useBranchReport
```

---

## Phase 4: Mobile App (sa-mb)

### Prompt 4.1 — Khởi tạo Flutter Project

```
Khởi tạo sa-mb — Employee Mobile App cho HDBBank Smart Attendance:

Tech stack:
- Flutter 3.16+ / Dart 3.2+
- State management: flutter_bloc 8.1.3
- HTTP: Dio 5.4.0 với JWT interceptor
- GPS: geolocator 11.0
- WiFi: network_info_plus 5.0
- Device info: device_info_plus 10.1
- Secure storage: flutter_secure_storage 9.0
- Push: firebase_messaging 14.7 + flutter_local_notifications 17.0
- UI: Material Design 3, google_fonts (Inter), shimmer 3.0
- Others: cached_network_image, flutter_svg, intl, connectivity_plus,
  permission_handler

Tạo:
1. pubspec.yaml với tất cả dependencies
2. lib/main.dart — entry point
3. lib/app.dart — root widget, DI container, routing
4. Folder structure theo Clean Architecture:
   core/, data/, domain/, presentation/
```

### Prompt 4.2 — Core Layer (Constants, Network, Theme)

```
Tạo lib/core/:

1. constants/api_constants.dart:
   - baseUrl configurable: localhost (iOS), 10.0.2.2 (Android emulator),
     LAN IP (physical device), production URL
   - API path constants

2. constants/app_constants.dart:
   - Token storage keys
   - Timeout values
   - Attendance status codes

3. network/api_client.dart:
   - Dio instance với baseUrl, timeout 30s
   - JWT interceptor: auto-add Bearer token từ secure storage
   - 401 response → clear tokens → navigate to login
   - Error handling: convert DioException → AppError

4. theme/app_colors.dart:
   - HDBank brand colors (primary, secondary)
   - Status colors (green/orange/red/blue)
   - Background, surface, text colors

5. theme/app_theme.dart:
   - Material Design 3 ThemeData
   - Light theme (dark theme ready)
   - Typography: Inter font via google_fonts

6. utils/date_utils.dart:
   - Format date/time Vietnamese locale
   - Calculate week/month ranges
```

### Prompt 4.3 — Data Layer (Models, Repositories, Services)

```
Tạo lib/data/:

1. models/ — Dart data classes matching backend:
   - api_response_model.dart: ApiResponse<T>, ApiError
   - user_model.dart: UserModel.fromJson/toJson (match entity.User)
   - branch_model.dart: BranchModel (match entity.Branch)
   - attendance_model.dart: AttendanceModel (match entity.AttendanceLog)
   - login_response_model.dart: LoginResponseModel (tokens + user)

2. repositories/ — API implementations:
   - auth_repository_impl.dart: login, getMe, changePassword
     (save tokens to secure storage on login)
   - attendance_repository_impl.dart: checkIn, checkOut, getTodayRecord,
     getHistory (paginated)

3. services/ — Platform services:
   - location_service.dart: getCurrentPosition, checkGeofence (compare with
     branch GPS config), detectMockedLocation (position.isMocked)
   - wifi_service.dart: getSSID, getBSSID (need location permission on Android 12+)
   - device_service.dart: getDeviceId (persistent), getDeviceModel, getAppVersion
   - security_service.dart: detectVPN (scan network interfaces for tun/tap/ppp/ipsec/utun),
     detectFakeGPS (from geolocator), runAllChecks → SecurityResult
```

### Prompt 4.4 — Domain Layer

```
Tạo lib/domain/:

1. repositories/auth_repository.dart:
   - Abstract class AuthRepository
   - login(email, password) → LoginResponseModel
   - getMe() → UserModel
   - changePassword(oldPassword, newPassword) → void
   - logout() → void
   - isAuthenticated() → bool

2. repositories/attendance_repository.dart:
   - Abstract class AttendanceRepository
   - checkIn(CheckInRequest) → AttendanceModel
   - checkOut(CheckOutRequest) → AttendanceModel
   - getTodayRecord() → AttendanceModel?
   - getHistory(period, page, limit) → PaginatedResponse<AttendanceModel>
```

### Prompt 4.5 — Presentation: BLoCs

```
Tạo lib/presentation/blocs/:

1. auth/:
   - auth_event.dart: AppStarted, LoginRequested(email, password),
     LogoutRequested
   - auth_state.dart: AuthInitial, AuthLoading, Authenticated(user),
     Unauthenticated, AuthError(message)
   - auth_bloc.dart:
     - AppStarted → check secure storage → Authenticated / Unauthenticated
     - LoginRequested → call repository.login → save tokens → Authenticated
     - LogoutRequested → clear tokens → Unauthenticated

2. attendance/:
   - attendance_event.dart: LoadTodayRecord, CheckInRequested(method, wifiSSID,
     wifiBSSID, latitude, longitude, deviceID, deviceModel, isFakeGPS, isVPN),
     CheckOutRequested(...), LoadHistory(period)
   - attendance_state.dart: AttendanceInitial, AttendanceLoading,
     TodayLoaded(record), CheckInSuccess(record), CheckOutSuccess(record),
     HistoryLoaded(records), AttendanceError(message)
   - attendance_bloc.dart:
     - CheckInRequested:
       1. Run SecurityService.runAllChecks()
       2. If fraud detected → emit error
       3. Call repository.checkIn()
       4. Emit CheckInSuccess
     - LoadHistory → call repository.getHistory → emit HistoryLoaded
```

### Prompt 4.6 — Presentation: Screens

```
Tạo lib/presentation/screens/:

1. login_screen.dart:
   - Email + password TextFormField
   - Form validation (email format, password ≥ 6)
   - Login button → dispatch LoginRequested to AuthBloc
   - BlocListener: AuthError → show SnackBar, Authenticated → navigate Home
   - Loading overlay khi AuthLoading

2. home_screen.dart:
   - BottomNavigationBar 3 tabs: Hôm nay, Lịch sử, Tài khoản
   - Tab "Hôm nay":
     - Greeting: "Xin chào, [Tên]"
     - Branch info card
     - Today's attendance card (check-in/out time, status)
     - Check-in / Check-out button (disable theo trạng thái)
     - Navigate check_in_screen khi tap

3. check_in_screen.dart:
   - 2 method buttons: WiFi, GPS
   - WiFi: request permission → scan → show SSID/BSSID
   - GPS: request permission → get location → show on map/text
   - Anti-fraud notice: "Hệ thống sẽ kiểm tra VPN, Fake GPS, Device ID"
   - Submit button → run security checks → dispatch CheckInRequested
   - BlocListener: success → toast + pop back, error → show message

4. history_screen.dart:
   - SegmentedButton: Tuần này / Tháng này / Tùy chọn
   - ListView of AttendanceCard widgets
   - Pull-to-refresh
   - Empty state message
```

### Prompt 4.7 — Presentation: Widgets

```
Tạo lib/presentation/widgets/:

1. attendance_card.dart:
   - Card hiển thị 1 ngày chấm công:
     - Ngày (formatted Vietnamese)
     - Giờ vào — Giờ ra
     - Method icon (WiFi / GPS)
     - StatusBadge
     - Số giờ làm
   - Ripple effect khi tap (xem chi tiết)

2. status_badge.dart:
   - Container với rounded corners
   - Color mapping: present→green, late→orange, absent→red, half_day→blue
   - Text label tiếng Việt

3. loading_overlay.dart:
   - Full-screen semi-transparent overlay
   - CircularProgressIndicator center
   - Prevent user interaction khi loading
```

---

## Phase 5: Tài liệu & Hoàn thiện

### Prompt 5.1 — README cho từng subproject

```
Tạo README.md cho từng subproject:

1. sa-api/README.md:
   - Kiến trúc Clean Architecture (4 layers diagram)
   - Source code structure
   - Features list (RBAC, dual validation, anti-fraud, rate limiting)
   - Setup & installation (Go, PostgreSQL, Redis, Docker)
   - Database migrations
   - API reference (tất cả endpoints)
   - Performance optimizations
   - Environment variables

2. sa-web/README.md:
   - Features theo từng page
   - Tech stack chi tiết
   - Architecture & data flow diagram
   - Folder structure
   - Setup (npm install, env, dev server)
   - Page-by-page API mapping
   - Development guidelines

3. sa-mb/README.md:
   - System requirements
   - Environment setup (Flutter SDK)
   - Project initialization
   - API backend configuration
   - Firebase setup (optional)
   - Source code structure
   - Architecture (Clean Architecture + BLoC)
   - Dependencies list
   - Main screens description
   - Development notes
```

### Prompt 5.2 — README tổng (Root)

```
Tạo README.md ở root — tài liệu tổng quan hệ thống:

Bao gồm:
1. Tổng quan kiến trúc — diagram ASCII art:
   Client Layer (sa-web + sa-mb) → sa-api → PostgreSQL + Redis

2. Tech stack tóm tắt — bảng đầy đủ với lý do chọn

3. Cơ chế Check-in Anti-fraud — phần phức tạp nhất:
   - Sơ đồ luồng 3 lớp validation (Rate Limit → Anti-fraud → Location → Business Rules)
   - Chi tiết từng lớp
   - Fraud scoring mechanism

4. Kiến trúc Multi-branch & Data Isolation:
   - Branch scoping cho Manager
   - Indexed queries
   - Geofencing config

5. Hệ thống phân quyền RBAC:
   - Ma trận quyền Admin/Manager/Employee
   - Middleware chain

6. Chiến lược Scaling 5.000 nhân viên đồng thời:
   - Connection pooling
   - Redis caching strategy
   - Database indexing
   - Rate limiting

7. Git Flow & CI/CD

8. Folder Structure tổng

9. Cài đặt & Chạy Project — step by step cho cả 3 subprojects
```

---

## Ghi chú

- Mỗi prompt ở trên đã được điều chỉnh để phản ánh đúng thứ tự xây dựng logic của dự án.
- Thực tế có thể đã có các prompt bổ sung để debug, refine, hoặc thêm edge cases.
- CLAUDE.md được tạo đầu tiên để làm "nguồn sự thật" cho toàn bộ quá trình phát triển.
