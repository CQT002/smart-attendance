# sa-web — Admin Portal

> Giao diện quản trị hệ thống **Smart Attendance** của HDBBank.
> Quản lý **100 chi nhánh — 5.000 nhân viên** thông qua dashboard realtime, báo cáo chấm công và CRUD đầy đủ.

**Stack:** Next.js 15 (App Router) · TypeScript · Tailwind CSS · shadcn/ui · TanStack Query v5

---

## Mục lục

1. [Tính năng](#1-tính-năng)
2. [Tech Stack chi tiết](#2-tech-stack-chi-tiết)
3. [Kiến trúc & Luồng dữ liệu](#3-kiến-trúc--luồng-dữ-liệu)
4. [Folder Structure](#4-folder-structure)
5. [Cài đặt & Chạy](#5-cài-đặt--chạy)
6. [Biến môi trường](#6-biến-môi-trường)
7. [Các trang & API tương ứng](#7-các-trang--api-tương-ứng)
8. [Hướng dẫn phát triển](#8-hướng-dẫn-phát-triển)

---

## 1. Tính năng

| Trang | Mô tả |
|---|---|
| **Login** | Xác thực JWT, validation Zod, hiện/ẩn mật khẩu, xử lý lỗi từ API |
| **Dashboard** | 6 KPI cards, Pie chart phân bố trạng thái, Bar chart theo chi nhánh, bảng realtime tự cập nhật mỗi 5 phút |
| **Quản lý Chi nhánh** | CRUD đầy đủ, cấu hình GPS Geofencing (lat/lng/radius), tìm kiếm, phân trang |
| **Quản lý Nhân viên** | Danh sách 5.000 nhân viên, lọc theo chi nhánh/vai trò, gán chi nhánh, reset mật khẩu |
| **Dữ liệu Chấm công** | DataTable với filter ngày/chi nhánh/trạng thái, hiển thị phương thức WiFi/GPS, cờ gian lận |
| **Báo cáo** | Báo cáo theo nhân viên và chi nhánh, chọn kỳ daily/weekly/monthly/custom, biểu đồ so sánh |

**Trải nghiệm người dùng:**
- Loading Shimmer thay skeleton trắng trơn trên mọi trạng thái chờ API
- Toast notification (Sonner) cho tất cả thao tác thành công / lỗi
- Sidebar có thể thu gọn, responsive trên mọi kích thước màn hình
- Pagination thông minh với hiển thị range kết quả

---

## 2. Tech Stack chi tiết

| Thư viện | Phiên bản | Mục đích |
|---|---|---|
| `next` | 15.3.0 | App Router, SSR, route groups |
| `react` | 19 | UI framework |
| `typescript` | 5 | Type safety |
| `tailwindcss` | 3.4 | Utility-first CSS |
| `@radix-ui/*` | latest | Accessible UI primitives |
| `class-variance-authority` | 0.7 | Variant-based component styling |
| `@tanstack/react-query` | 5.62 | Server-state management, caching |
| `axios` | 1.7 | HTTP client với interceptors |
| `react-hook-form` | 7.54 | Form state, ít re-render |
| `zod` | 3.23 | Schema validation |
| `@hookform/resolvers` | 3.9 | Bridge RHF ↔ Zod |
| `recharts` | 2.13 | Biểu đồ Bar, Pie |
| `date-fns` | 4.1 | Format ngày giờ tiếng Việt |
| `sonner` | 1.7 | Toast notifications |
| `js-cookie` | 3.0 | JWT cookie management |
| `lucide-react` | 0.468 | Icon set |

---

## 3. Kiến trúc & Luồng dữ liệu

### Tổng quan

```
Browser
  │
  ├── app/(auth)/login          → Không cần xác thực
  │
  └── app/(admin)/              → Route group (yêu cầu JWT)
        ├── layout.tsx          → Auth guard + Sidebar layout
        ├── dashboard/
        ├── branches/
        ├── users/
        ├── attendance/
        └── reports/
```

### Luồng dữ liệu (Data Flow)

```
Page Component
    │  gọi hook
    ▼
hooks/use-*.ts          ← TanStack Query (cache, stale-while-revalidate)
    │  gọi service
    ▼
services/*.service.ts   ← Abstraction layer, xử lý response shape
    │  gọi apiClient
    ▼
lib/api-client.ts       ← Axios instance + JWT interceptor + 401 handler
    │  HTTP request
    ▼
Backend API (sa-api)    ← Go/Echo, port 8080
```

### Authentication Flow

```
1. User nhập email/password → POST /admin/auth/login
2. Backend trả về { access_token, refresh_token, user }
3. authService.login():
   - Set cookie "access_token" (1 ngày)
   - Set cookie "refresh_token" (7 ngày)
   - Lưu user profile vào localStorage
4. Mọi request tiếp theo: Axios interceptor tự đính Bearer token
5. 401 response: Xóa cookies + redirect /login
6. Logout: queryClient.clear() + xóa cookies/localStorage
```

### Route Protection

`app/(admin)/layout.tsx` kiểm tra `isAuthenticated()` (có cookie `access_token`) mỗi khi mount. Nếu không hợp lệ → `router.replace("/login")`.

---

## 4. Folder Structure

```
sa-web/
│
├── app/                              # Next.js App Router
│   ├── layout.tsx                    # Root layout: font Inter, metadata, Providers
│   ├── globals.css                   # CSS variables (light/dark mode) + shimmer keyframe
│   ├── page.tsx                      # / → redirect /dashboard
│   ├── providers.tsx                 # QueryClient (staleTime 30s) + Sonner Toaster
│   │
│   ├── (auth)/                       # Route group — không có sidebar
│   │   ├── layout.tsx
│   │   └── login/
│   │       └── page.tsx              # Form đăng nhập, Zod validation
│   │
│   └── (admin)/                      # Route group — có Sidebar + auth guard
│       ├── layout.tsx                # isAuthenticated check + Sidebar wrapper
│       ├── dashboard/
│       │   └── page.tsx              # KPI cards, Pie/Bar charts, branch table
│       ├── branches/
│       │   └── page.tsx              # CRUD chi nhánh + GPS config
│       ├── users/
│       │   └── page.tsx              # Danh sách nhân viên + CRUD
│       ├── attendance/
│       │   └── page.tsx              # DataTable logs chấm công
│       └── reports/
│           └── page.tsx              # Báo cáo NV/Chi nhánh + biểu đồ
│
├── components/
│   ├── ui/                           # shadcn/ui primitives (tự viết, không dùng CLI)
│   │   ├── badge.tsx                 # Có thêm variant: success, warning, info
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── dialog.tsx
│   │   ├── input.tsx
│   │   ├── label.tsx
│   │   ├── select.tsx
│   │   ├── skeleton.tsx              # Dùng shimmer CSS animation
│   │   └── table.tsx
│   │
│   ├── layout/
│   │   ├── sidebar.tsx               # Collapsible sidebar, active link highlight
│   │   └── header.tsx                # Tên trang + current user info
│   │
│   ├── shared/                       # Dùng lại trên nhiều trang
│   │   ├── pagination.tsx            # Điều hướng trang với range display
│   │   ├── status-badge.tsx          # StatusBadge, RoleBadge, ActiveBadge
│   │   └── data-table-skeleton.tsx   # Skeleton đúng shape của table
│   │
│   ├── branches/
│   │   └── branch-form-dialog.tsx    # Form Create/Edit chi nhánh + GPS fields
│   │
│   └── users/
│       ├── user-form-dialog.tsx      # Form Create/Edit nhân viên
│       └── reset-password-dialog.tsx # Dialog reset mật khẩu
│
├── hooks/                            # TanStack Query wrappers
│   ├── use-auth.ts                   # useCurrentUser, useLogin, useLogout
│   ├── use-users.ts                  # useUsers, useCreateUser, useUpdateUser...
│   ├── use-branches.ts               # useBranches, useActiveBranches, useCreateBranch...
│   ├── use-attendance.ts             # useAttendanceLogs, useAttendanceSummary
│   └── use-reports.ts                # useDashboardStats (refetch 5 phút), useTodayBranchStats...
│
├── services/                         # API service layer
│   ├── auth.service.ts               # login, getMe, changePassword, logout
│   ├── user.service.ts               # getList, getById, create, update, delete, resetPassword
│   ├── branch.service.ts             # getList, getActive, getById, create, update, delete
│   ├── attendance.service.ts         # getList, getById, getSummary
│   └── report.service.ts             # getDashboardStats, getTodayBranchStats, getAttendanceReport...
│
├── lib/
│   ├── api-client.ts                 # Axios instance, JWT interceptor, 401 redirect, buildQueryString
│   ├── auth.ts                       # isAuthenticated, getStoredUser, setStoredUser, clearStoredUser
│   └── utils.ts                      # cn(), formatDate, formatDateTime, formatTime, formatPercent, formatHours
│
├── types/                            # TypeScript interfaces — đồng bộ với Go structs
│   ├── api.ts                        # ApiResponse<T>, PaginationMeta, PaginatedResponse<T>
│   ├── auth.ts                       # UserRole, LoginRequest, LoginResponse, AuthUser
│   ├── user.ts                       # User, CreateUserRequest, UpdateUserRequest, UserFilter
│   ├── branch.ts                     # Branch, WifiNetwork, CreateBranchRequest, BranchFilter
│   ├── attendance.ts                 # AttendanceLog, AttendanceStatus, AttendanceFilter
│   ├── report.ts                     # DashboardStats, BranchTodayStats, ReportFilter
│   └── index.ts                      # Re-export tất cả
│
├── .env.local                        # NEXT_PUBLIC_API_URL (không commit)
├── components.json                   # shadcn/ui config
├── next.config.ts
├── tailwind.config.ts                # CSS variables, shimmer animation
├── tsconfig.json                     # strict mode, path alias @/*
└── package.json
```

---

## 5. Cài đặt & Chạy

### Yêu cầu

- **Node.js** ≥ 20
- **npm** ≥ 10 (hoặc pnpm/yarn)
- Backend `sa-api` đang chạy tại `http://localhost:8080`

### Các bước

**Bước 1 — Cài dependencies**

```bash
cd sa-web
npm install
```

**Bước 2 — Cấu hình environment**

```bash
# File .env.local đã có sẵn với giá trị mặc định
# Kiểm tra và chỉnh nếu backend chạy cổng khác
cat .env.local
# NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

**Bước 3 — Khởi động development server**

```bash
npm run dev
# → http://localhost:3000
```

Server dùng **Turbopack** (Next.js 15), thời gian cold start < 1 giây.

**Bước 4 — Đăng nhập**

Mở `http://localhost:3000`, trình duyệt tự redirect sang `/login`.

```
Email:    admin@hdbank.com.vn
Password: Admin@123
```

> Tài khoản seed được tạo bởi migration `sa-api/migrations/001_init_schema.sql`.

---

### Build production

```bash
# Build
npm run build

# Preview production build locally
npm start
# → http://localhost:3000
```

### Lint

```bash
npm run lint
```

---

## 6. Biến môi trường

| Biến | Bắt buộc | Mặc định | Mô tả |
|---|:---:|---|---|
| `NEXT_PUBLIC_API_URL` | ✅ | `http://localhost:8080/api/v1` | Base URL của backend API |

> Prefix `NEXT_PUBLIC_` để biến được expose ra browser bundle. Không đặt secret vào đây.

**Production** — tạo file `.env.production.local`:

```bash
NEXT_PUBLIC_API_URL=https://api.hdbank.vn/api/v1
```

---

## 7. Các trang & API tương ứng

### Login — `/login`

| Mục | Chi tiết |
|---|---|
| Validation | Zod: email format, password ≥ 6 ký tự |
| API | `POST /api/v1/admin/auth/login` |
| Lưu trạng thái | Cookie `access_token` (1 ngày), `refresh_token` (7 ngày), `localStorage["current_user"]` |
| Redirect | Thành công → `/dashboard` |

### Dashboard — `/dashboard`

| Mục | Chi tiết |
|---|---|
| API | `GET /admin/reports/dashboard` — 6 KPI metrics |
| API | `GET /admin/reports/today?branch_id=&page=` — stats từng chi nhánh |
| Realtime | TanStack Query `refetchInterval: 5 phút` — không cần F5 |
| Filter | Dropdown chi nhánh (dữ liệu từ `GET /admin/branches/active`) |
| Charts | `recharts` — PieChart (present/late/absent), BarChart (top 8 chi nhánh) |
| Loading | Skeleton shimmer cho cards và charts trong lúc fetch |

### Quản lý Chi nhánh — `/branches`

| Mục | Chi tiết |
|---|---|
| API CRUD | `GET/POST /admin/branches`, `PUT/DELETE /admin/branches/:id` |
| Tìm kiếm | Query param `?search=` — debounce khi nhấn Enter hoặc click icon |
| GPS Config | Form fields: Latitude, Longitude, Radius (50–5000m) |
| Xoá | Soft-delete (backend đặt `is_active = false`) với confirm dialog |
| Phân quyền | Nút Thêm/Sửa/Xoá chỉ hiện với `role === "admin"` |

### Quản lý Nhân viên — `/users`

| Mục | Chi tiết |
|---|---|
| API | `GET /admin/users?branch_id=&role=&search=&page=` |
| Filter | Chi nhánh + Vai trò + Full-text search (tên/email/mã NV) |
| Tạo mới | Form đầy đủ: mã NV, vai trò, email, mật khẩu, chi nhánh, phòng ban |
| Reset password | Dialog riêng `POST /admin/users/:id/reset-password` |
| Phân trang | `limit: 20`, hiển thị tổng số kết quả |

### Dữ liệu Chấm công — `/attendance`

| Mục | Chi tiết |
|---|---|
| API | `GET /admin/attendance?date_from=&date_to=&branch_id=&status=&page=` |
| Filter | Khoảng ngày (date picker) + Chi nhánh + Trạng thái |
| Trạng thái | `present` / `late` / `early_leave` / `absent` / `half_day` — mỗi loại badge màu riêng |
| Gian lận | Cột "Gian lận" hiện badge đỏ `GPS giả` hoặc `VPN` nếu `is_fake_gps` hoặc `is_vpn = true` |
| Phương thức | Check-in method `WIFI` / `GPS` hiển thị dưới giờ vào |

### Báo cáo — `/reports`

| Mục | Chi tiết |
|---|---|
| API | `GET /admin/reports/attendance` — báo cáo theo nhân viên |
| API | `GET /admin/reports/branches` — báo cáo theo chi nhánh |
| Kỳ báo cáo | `daily` / `weekly` / `monthly` / `custom` (custom hiện thêm date range) |
| Filter | Chi nhánh + Phòng ban + Kỳ |
| Tab | Chuyển đổi giữa "Theo nhân viên" và "Theo chi nhánh" |
| Biểu đồ | BarChart tỷ lệ chuyên cần vs đúng giờ theo chi nhánh |
| Màu KPI | `≥ 90%` xanh · `70–89%` vàng · `< 70%` đỏ |

---

## 8. Hướng dẫn phát triển

### Thêm trang mới

1. Tạo file trong `app/(admin)/<tên-trang>/page.tsx`
2. Thêm nav item vào `components/layout/sidebar.tsx`
3. Tạo service trong `services/<tên>.service.ts`
4. Tạo hook trong `hooks/use-<tên>.ts`
5. Khai báo types trong `types/<tên>.ts` và export qua `types/index.ts`

### Thêm UI component mới

Đặt vào `components/ui/` nếu là primitive tái sử dụng, `components/shared/` nếu có business logic nhẹ (như `StatusBadge`), hoặc `components/<feature>/` nếu gắn với một tính năng cụ thể.

### Quy tắc gọi API

Không gọi `apiClient` trực tiếp từ component. Luôn đi qua service → hook:

```typescript
// ✅ Đúng
const { data, isLoading } = useUsers({ branch_id: 1, page: 1 });

// ❌ Sai — gọi thẳng từ component
const res = await apiClient.get("/admin/users");
```

### Xử lý lỗi

- Mutation hooks đã có `onError: () => toast.error(...)` sẵn.
- Lỗi network / 5xx được Axios interceptor bắt, log vào console.
- Lỗi 401 tự động redirect `/login`.
- Form validation hiện lỗi inline qua `react-hook-form` + Zod.

### Cấu trúc API Response

Tất cả response từ backend theo chuẩn:

```typescript
// types/api.ts
interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: { code: string; message: string };
  meta?: PaginationMeta;  // Chỉ có ở list endpoints
}

interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}
```

Service layer unwrap `res.data.data` trước khi trả về hook, giữ component sạch.

### CSS Variables & Dark Mode

Màu sắc được định nghĩa qua HSL CSS variables trong `globals.css`:

```css
:root {
  --primary: 221.2 83.2% 53.3%;    /* Blue */
  --destructive: 0 84.2% 60.2%;    /* Red */
  --sidebar-background: 222.2 84% 4.9%; /* Dark navy */
  ...
}
```

Sidebar dùng `--sidebar-*` variables độc lập để có thể có theme riêng mà không ảnh hưởng main content.

### Shimmer Loading

Tất cả trạng thái loading dùng component `Skeleton` với class `.shimmer`:

```tsx
// Cho table
<DataTableSkeleton columns={6} rows={8} />

// Cho card đơn lẻ
<Skeleton className="h-8 w-20" />

// Cho chart
<Skeleton className="h-64 w-full" />
```

Animation được định nghĩa trong `globals.css`, không phụ thuộc thư viện ngoài.
