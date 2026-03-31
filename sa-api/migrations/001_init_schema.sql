-- =============================================================
-- Smart Attendance - Initial Schema Migration
-- Tối ưu hóa cho 5000 users và 100 chi nhánh
-- =============================================================

-- Bảng chi nhánh
CREATE TABLE IF NOT EXISTS branches (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(200) NOT NULL,
    address     VARCHAR(500),
    phone       VARCHAR(20),
    email       VARCHAR(100),
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_branches_code UNIQUE (code)
);

-- Index tìm kiếm theo trạng thái active (dùng nhiều nhất)
CREATE INDEX idx_branches_is_active ON branches (is_active);

-- Bảng người dùng
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    branch_id     INTEGER      REFERENCES branches(id) ON DELETE SET NULL,
    employee_code VARCHAR(50)  NOT NULL,
    name          VARCHAR(200) NOT NULL,
    email         VARCHAR(100) NOT NULL,
    phone         VARCHAR(20),
    password      VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'employee' CHECK (role IN ('admin', 'manager', 'employee')),
    department    VARCHAR(100),
    position      VARCHAR(100),
    avatar_url    VARCHAR(500),
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_users_email         UNIQUE (email),
    CONSTRAINT uq_users_employee_code UNIQUE (employee_code)
);

-- Composite index cho query phổ biến: lấy nhân viên theo chi nhánh + trạng thái
CREATE INDEX idx_users_branch_active  ON users (branch_id, is_active);
-- Index tìm theo role
CREATE INDEX idx_users_role           ON users (role);
-- Index tìm theo phòng ban
CREATE INDEX idx_users_department     ON users (department);

-- Bảng ca làm việc
CREATE TABLE IF NOT EXISTS shifts (
    id           SERIAL PRIMARY KEY,
    branch_id    INTEGER      NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name         VARCHAR(100) NOT NULL,
    start_time   VARCHAR(5)   NOT NULL,  -- HH:MM
    end_time     VARCHAR(5)   NOT NULL,  -- HH:MM
    late_after   INTEGER      NOT NULL DEFAULT 15,    -- phút
    early_before INTEGER      NOT NULL DEFAULT 15,    -- phút
    work_hours   DECIMAL(5,2),
    is_default   BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shifts_branch_active ON shifts (branch_id, is_active);

-- Bảng cấu hình WiFi
CREATE TABLE IF NOT EXISTS wifi_configs (
    id          SERIAL PRIMARY KEY,
    branch_id   INTEGER      NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    ssid        VARCHAR(100) NOT NULL,
    bssid       VARCHAR(50),              -- MAC address của router
    description VARCHAR(200),
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index tìm theo BSSID khi xác thực check-in
CREATE INDEX idx_wifi_configs_branch_active ON wifi_configs (branch_id, is_active);
CREATE INDEX idx_wifi_configs_bssid         ON wifi_configs (bssid) WHERE bssid IS NOT NULL;

-- Bảng cấu hình GPS Geofencing
CREATE TABLE IF NOT EXISTS gps_configs (
    id          SERIAL PRIMARY KEY,
    branch_id   INTEGER         NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name        VARCHAR(100),
    latitude    DECIMAL(10, 8)  NOT NULL,
    longitude   DECIMAL(11, 8)  NOT NULL,
    radius      DECIMAL(8, 2)   NOT NULL,  -- bán kính tính bằng mét
    description VARCHAR(200),
    is_active   BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gps_configs_branch_active ON gps_configs (branch_id, is_active);

-- Bảng log chấm công (bảng lớn nhất, cần index kỹ)
CREATE TABLE IF NOT EXISTS attendance_logs (
    id                SERIAL PRIMARY KEY,
    user_id           INTEGER          NOT NULL REFERENCES users(id),
    branch_id         INTEGER          NOT NULL REFERENCES branches(id),
    shift_id          INTEGER          REFERENCES shifts(id),
    date              DATE             NOT NULL,

    -- Check-in
    check_in_time     TIMESTAMPTZ,
    check_in_lat      DECIMAL(10, 8),
    check_in_lng      DECIMAL(11, 8),
    check_in_method   VARCHAR(10)      CHECK (check_in_method IN ('wifi', 'gps')),
    check_in_ssid     VARCHAR(100),
    check_in_bssid    VARCHAR(50),

    -- Check-out
    check_out_time    TIMESTAMPTZ,
    check_out_lat     DECIMAL(10, 8),
    check_out_lng     DECIMAL(11, 8),
    check_out_method  VARCHAR(10)      CHECK (check_out_method IN ('wifi', 'gps')),
    check_out_ssid    VARCHAR(100),
    check_out_bssid   VARCHAR(50),

    -- Device & security
    device_id         VARCHAR(200),
    device_model      VARCHAR(100),
    ip_address        VARCHAR(45),
    app_version       VARCHAR(20),

    -- Anti-fraud flags
    is_fake_gps       BOOLEAN          NOT NULL DEFAULT FALSE,
    is_vpn            BOOLEAN          NOT NULL DEFAULT FALSE,
    fraud_note        VARCHAR(500),

    -- Calculated
    status            VARCHAR(20)      CHECK (status IN ('present', 'late', 'early_leave', 'absent', 'half_day')),
    work_hours        DECIMAL(5, 2)    NOT NULL DEFAULT 0,
    overtime          DECIMAL(5, 2)    NOT NULL DEFAULT 0,
    note              VARCHAR(500),

    created_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW(),

    -- Một user chỉ có một bản ghi chấm công mỗi ngày
    CONSTRAINT uq_attendance_user_date UNIQUE (user_id, date)
);

-- === Critical Indexes cho Performance với 5000 users ===

-- Index tìm chấm công theo user và ngày (query thường xuyên nhất)
CREATE INDEX idx_attendance_user_date     ON attendance_logs (user_id, date DESC);
-- Index tìm theo chi nhánh và ngày (cho manager dashboard)
CREATE INDEX idx_attendance_branch_date   ON attendance_logs (branch_id, date DESC);
-- Index lọc theo trạng thái (cho báo cáo)
CREATE INDEX idx_attendance_status        ON attendance_logs (status, date DESC);
-- Index tìm bản ghi gian lận
CREATE INDEX idx_attendance_fraud         ON attendance_logs (user_id, is_fake_gps, is_vpn, created_at DESC)
    WHERE is_fake_gps = TRUE OR is_vpn = TRUE;
-- Index theo device ID để phát hiện thiết bị lạ
CREATE INDEX idx_attendance_device        ON attendance_logs (device_id, user_id);

-- =============================================================
-- Seed data: Admin account mặc định
-- Mật khẩu: Admin@123 (đã hash bcrypt)
-- =============================================================
INSERT INTO users (employee_code, name, email, password, role, is_active)
VALUES (
    'ADMIN001',
    'System Administrator',
    'admin@hdbank.com.vn',
    '$2a$10$ZSZnC8n7hO8awy2PHsSrSOY8bfwYHCpF5/yqT7yuCNK8/gcvy0CAW',
    'admin',
    TRUE
) ON CONFLICT (email) DO NOTHING;
