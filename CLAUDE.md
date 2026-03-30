# CLAUDE.md — HDBBank Smart Attendance

Tài liệu này định nghĩa các quy tắc bắt buộc cho AI khi làm việc với project này.
Đây là nguồn sự thật duy nhất — mọi hướng dẫn ở đây đều OVERRIDE hành vi mặc định.

---

## 1. Tech Stack

| Layer | Công nghệ |
|---|---|
| Language | Go 1.22 |
| HTTP Framework | Echo v4 (`github.com/labstack/echo/v4`) |
| ORM | GORM v2 (`gorm.io/gorm`) + driver `gorm.io/driver/postgres` |
| Database | PostgreSQL 16 |
| Cache | Redis 7 (`github.com/redis/go-redis/v9`) |
| Auth | JWT HS256 (`github.com/golang-jwt/jwt/v5`) |
| Config | Viper (`github.com/spf13/viper`) + `config/config.yaml` |
| Logging | `log/slog` (stdlib, không dùng thư viện ngoài) |
| Password | `golang.org/x/crypto/bcrypt` cost 10 |

**Không được tự ý thêm dependency mới** mà không có lý do rõ ràng.
Trước khi thêm package, hỏi xem stdlib hoặc các dep hiện tại có đáp ứng được không.

---

## 2. Cấu trúc thư mục — Clean Architecture

```
sa-api/
├── cmd/server/main.go              # Entrypoint, Dependency Injection
├── config/
│   ├── config.go                   # Viper loader, struct Config
│   └── config.yaml                 # Giá trị config (dev). KHÔNG commit secret thật
├── internal/
│   ├── domain/                     # Layer trong cùng — KHÔNG import layer ngoài
│   │   ├── entity/                 # GORM models, business types
│   │   ├── repository/             # Repository interfaces
│   │   └── usecase/                # Usecase interfaces + request/response types
│   ├── repository/                 # Implementations của domain/repository (PostgreSQL)
│   ├── usecase/
│   │   ├── attendance/main.go
│   │   ├── branch/main.go
│   │   ├── report/main.go
│   │   └── user/main.go
│   ├── handler/
│   │   ├── admin/                  # HTTP handlers cho Admin Portal
│   │   └── user/                   # HTTP handlers cho Employee App
│   ├── middleware/                 # JWT auth, rate limiter, request logger
│   ├── infrastructure/
│   │   ├── cache/                  # Redis client + Cache interface
│   │   ├── database/               # GORM setup, AutoMigrate
│   │   └── logger/                 # slog setup (JSON/Text)
│   └── server/router.go            # Echo route registration
├── pkg/
│   ├── apperrors/                  # AppError, domain errors, ValidationError
│   ├── response/                   # Chuẩn hóa HTTP response
│   └── utils/                      # JWT helpers, geo utils, pagination
└── migrations/                     # SQL migration files
```

### Quy tắc dependency giữa các layer

```
handler → usecase interface → repository interface → entity
```

- `domain/` **không được** import bất kỳ package nào khác trong project.
- `repository/` chỉ import `domain/` và `infrastructure/`.
- `usecase/` chỉ import `domain/` và `infrastructure/cache`.
- `handler/` chỉ import `domain/usecase`, `middleware`, `pkg/`.
- **Không bao giờ** import implementation trực tiếp từ handler/usecase (phải qua interface).

---

## 3. Quy trình xử lý lỗi tập trung

### AppError — kiểu lỗi duy nhất

Mọi lỗi business logic phải trả về `*apperrors.AppError`:

```go
// pkg/apperrors/errors.go
type AppError struct {
    HTTPStatus int
    Code       string
    Message    string
    Err        error  // wrapped original error
}

// Tạo lỗi domain
var ErrUserNotFound = apperrors.New(404, "USER_NOT_FOUND", "Người dùng không tồn tại")

// Wrap lỗi infrastructure
return apperrors.Wrap(err, 500, "DB_ERROR", "Lỗi truy vấn database")
```

### Response chuẩn hóa

Handler **không bao giờ** tự format JSON — luôn dùng `pkg/response`:

```go
// Thành công
response.OK(c, data)
response.Created(c, data)
response.OKWithMessage(c, "Thành công", data)
response.Paginated(c, items, total, page, limit)

// Lỗi — tự detect AppError vs generic error
response.Error(c, err)  // → 4xx nếu AppError, 500 nếu không phải
```

### Không dùng panic để xử lý lỗi

Repository và Usecase luôn trả về `(T, error)` — không dùng `panic`.
Middleware `Recover()` của Echo chỉ là lưới an toàn cuối cùng.

---

## 4. Logging với slog

### Setup

```go
// development: Text format dễ đọc
// production:  JSON format để ingest vào log system
applogger.Setup(cfg.App.Env, cfg.App.Debug)
```

### Quy ước log level

| Level | Khi nào dùng |
|---|---|
| `slog.Info` | Sự kiện quan trọng: server start, user login, check-in thành công |
| `slog.Warn` | Bất thường nhưng không phải lỗi: rate limit, anti-fraud flag, cache miss |
| `slog.Error` | Lỗi cần xử lý: DB fail, Redis fail, token invalid |
| `slog.Debug` | Chi tiết debug — chỉ bật khi `APP_DEBUG=true` |

### Cách log đúng — luôn kèm context

```go
// ĐÚNG — structured logging với key-value
slog.Info("check-in successful",
    "user_id", req.UserID,
    "branch_id", req.BranchID,
    "status", attendLog.Status,
)

slog.Error("database query failed",
    "operation", "FindByEmail",
    "email", email,
    "error", err,
)

// SAI — không dùng fmt.Sprintf trong log
slog.Info(fmt.Sprintf("user %d checked in", userID))  // ❌
```

### Request logger middleware

Mọi HTTP request được log tự động bởi `middleware.RequestLogger()` với:
- `request_id` duy nhất mỗi request
- method, path, status, latency
- WARN nếu status 4xx, ERROR nếu 5xx

---

## 5. Conventional Commits

**Mọi commit phải tuân theo format:**

```
<type>(<scope>): <mô tả ngắn gọn bằng tiếng Anh>

[body tùy chọn]
```

### Allowed types

| Type | Dùng khi |
|---|---|
| `feat` | Thêm tính năng mới |
| `fix` | Sửa bug |
| `refactor` | Tái cấu trúc code, không thêm tính năng / không sửa bug |
| `perf` | Tối ưu hiệu năng |
| `test` | Thêm hoặc sửa tests |
| `docs` | Cập nhật tài liệu |
| `chore` | Cập nhật build, deps, config — không ảnh hưởng source code |
| `ci` | Thay đổi CI/CD pipeline |

### Allowed scopes

`auth`, `attendance`, `user`, `branch`, `report`, `middleware`, `config`, `db`, `cache`, `infra`, `handler`, `usecase`, `repository`

### Ví dụ commit hợp lệ

```
feat(attendance): add WiFi + GPS dual validation on check-in
fix(auth): prevent timing attack on login by constant-time compare
refactor(repository): flatten postgres/ subfolder into repository/
perf(report): replace N+1 query with single JOIN in GetBranchSummary
chore(config): migrate from .env to config.yaml with nested keys
```

### Commit message KHÔNG hợp lệ

```
fix: bug fix                          ❌ quá mơ hồ
feat: nhiều thứ linh tinh             ❌ tiếng Việt trong subject
update attendance handler             ❌ thiếu type
feat(attendance): Add check-in.       ❌ viết hoa + dấu chấm cuối
```

---

## 6. Code conventions

### Golang

- **Package names**: lowercase, một từ. `package user`, `package attendance` — không dùng `userUsecase`.
- **Error wrapping**: dùng `apperrors.Wrap()` khi bắt lỗi từ infrastructure, không dùng `fmt.Errorf`.
- **Context propagation**: mọi function chạm DB/Redis đều nhận `ctx context.Context` làm tham số đầu tiên.
- **Goroutine**: chỉ spawn goroutine cho fire-and-forget (ví dụ: `UpdateLastLogin`). Mọi trường hợp khác phải xử lý synchronous.
- **Interface**: định nghĩa interface ở phía consumer (domain layer), không ở phía implementor.

### GORM

- Luôn dùng `.WithContext(ctx)` trước mọi query.
- Không dùng `gorm:"column:..."` trừ khi tên field Go khác với convention snake_case.
- Dùng Preload thay vì raw JOIN khi load quan hệ 1-nhiều ≤ 100 records.
- Dùng raw SQL (`db.Raw(...)`) cho aggregate query phức tạp (dashboard, report).

### API

- Route prefix: `/api/v1/`
- Employee app routes: `/api/v1/attendance/*`, `/api/v1/auth/*`
- Admin portal routes: `/api/v1/admin/*`
- HTTP status codes: 200 OK, 201 Created, 204 No Content, 400 Validation, 401 Unauth, 403 Forbidden, 404 Not Found, 409 Conflict, 429 Rate Limit, 500 Internal.

---

## 7. Security

- **Không bao giờ** log password, JWT token, hoặc thông tin nhạy cảm.
- **Không commit** secret thật vào `config.yaml` — dùng placeholder `change-me-in-production`.
- `config.yaml` chứa giá trị dev-safe. Production secrets đến từ biến môi trường hoặc secret manager.
- Rate limiting bắt buộc trên login (10 req/15min per IP) và check-in (10 req/min per user).
- Anti-fraud check phải chạy trước mọi thao tác check-in/out.
