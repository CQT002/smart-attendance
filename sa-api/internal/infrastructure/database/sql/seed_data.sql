-- Seed data cho Smart Attendance — dữ liệu test mặc định
-- Chạy tự động lần đầu khi build source (qua seeder trong Go)
-- Password cho tất cả user: Admin@123 (admin), Password@123 (staff)

BEGIN;

-- ============================================================
-- 1. Branches
-- ============================================================
INSERT INTO branches (id, code, name, address, city, province, phone, email, is_active, created_at, updated_at)
VALUES
    (1, 'HCM001', 'Chi nhánh TP.HCM', '174 Phan Đăng Lưu', 'Tân Định', 'Thành phố Hồ Chí Minh', '0962503503', 'hcm001@hdbank.com.vn', true, '2026-03-31 22:27:48.799+07', '2026-04-02 14:02:53.236+07')
ON CONFLICT (id) DO NOTHING;

-- Cập nhật sequence branches
SELECT setval('branches_id_seq', (SELECT COALESCE(MAX(id), 1) FROM branches));

-- ============================================================
-- 2. Users
-- ============================================================
-- Admin: Admin@123
-- Staff: Password@123
INSERT INTO users (id, employee_code, name, email, phone, password, department, position, avatar_url, hired_at, last_login_at, created_at, updated_at, branch_id, role, is_active)
VALUES
    (1, 'ADMIN001', 'System Administrator', 'admin@hdbank.com.vn', '', '$2a$10$ZSZnC8n7hO8awy2PHsSrSOY8bfwYHCpF5/yqT7yuCNK8/gcvy0CAW', 'Phòng Kinh Doanh', 'Giám đốc hệ thống', '', NULL, NULL, '2026-04-03 16:02:20.830+07', '2026-03-31 21:46:32.406+07', NULL, 'admin', true),
    (2, 'NVA001', 'Nguyễn Văn A', 'nva@hdbank.com.vn', '0932564786', '$2a$10$RRhx4ZSUuBLemDHVu5QYjeyRlO5hDCVVhFn/507wOegckyE.Q8fs6', 'Công nghệ', 'Trưởng phòng', '', NULL, '2026-04-02 21:04:13.879+07', '2026-03-31 22:29:13.295+07', '2026-04-02 21:04:13.889+07', 1, 'manager', true),
    (3, 'TVB001', 'Trần Văn B', 'tvb@hdbank.com.vn', '0967334655', '$2a$10$8MZM9kW57EgsFYGLkdzKuOby4M297RXktKAwLvWHksq0Cv/zlsIn.', 'Công nghệ', 'Chuyên viên', '', NULL, '2026-04-03 14:35:42.635+07', '2026-03-31 22:30:09.583+07', '2026-04-03 14:35:42.640+07', 1, 'employee', true),
    (4, 'LVC001', 'Lê Văn Cường', 'lvc@hdbank.com.vn', '0902050301', '$2a$10$UTG9Tpou0s/WR81kMxW0YOgEY0cM9mL1LY/MXhgjTb6vLrFugvw0W', 'Công nghệ', 'Developer', '', NULL, NULL, '2026-04-03 14:58:38.640+07', '2026-04-03 14:58:38.640+07', 1, 'employee', true),
    (5, 'VVD001', 'Vũ Văn D', 'vvd@hdbank.com.vn', '0906885532', '$2a$10$Ozvb/VmYmEXLc8I/knDQzO0Qwt5rgBdTnKf7LsQHt83lDjvmQ1T0y', 'Công nghệ', 'Developer', '', NULL, '2026-04-03 16:00:56.142+07', '2026-04-03 15:06:47.592+07', '2026-04-03 16:00:56.144+07', 1, 'employee', true)
ON CONFLICT (id) DO NOTHING;

-- Cập nhật sequence users
SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 1) FROM users));

-- ============================================================
-- 3. WiFi Configs
-- ============================================================
INSERT INTO wifi_configs (id, branch_id, ssid, bssid, description, is_active, created_at, updated_at)
VALUES
    (1, 1, 'Long', '72:D0:1F:20:BE:26', 'Wifi nhà', true, '2026-03-31 22:27:48.824+07', '2026-03-31 22:27:48.824+07'),
    (2, 1, 'GIHub_7F', '22:AD:96:EE:C4:8E', 'Wifi lầu 7', true, '2026-04-02 14:02:51.037+07', '2026-04-02 14:02:51.037+07')
ON CONFLICT (id) DO NOTHING;

-- Cập nhật sequence wifi_configs
SELECT setval('wifi_configs_id_seq', (SELECT COALESCE(MAX(id), 1) FROM wifi_configs));

-- ============================================================
-- 4. GPS Configs
-- ============================================================
INSERT INTO gps_configs (id, branch_id, name, latitude, longitude, radius, description, is_active, created_at, updated_at)
VALUES
    (1, 1, 'Chi nhánh TP.HCM', 15.81020000, 105.87900000, 100.00, '', true, '2026-03-31 22:27:48.812+07', '2026-04-02 14:02:53.249+07')
ON CONFLICT (id) DO NOTHING;

-- Cập nhật sequence gps_configs
SELECT setval('gps_configs_id_seq', (SELECT COALESCE(MAX(id), 1) FROM gps_configs));

-- ============================================================
-- 5. Shifts
-- ============================================================
INSERT INTO shifts (id, branch_id, name, start_time, end_time, late_after, early_before, work_hours, is_default, is_active, created_at, updated_at)
VALUES
    (1, 1, 'Ca hành chính', '08:00', '17:00', 15, 15, 8.00, true, true, '2026-04-03 14:19:00.141+07', '2026-04-03 14:19:00.141+07')
ON CONFLICT (id) DO NOTHING;

-- Cập nhật sequence shifts
SELECT setval('shifts_id_seq', (SELECT COALESCE(MAX(id), 1) FROM shifts));

COMMIT;
