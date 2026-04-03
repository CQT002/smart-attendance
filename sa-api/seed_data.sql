-- Script SQL tạo 99 chi nhánh, WiFi, GPS, Ca làm việc và 1000 Nhân viên (có 99 Managers)
-- Khởi chạy trực tiếp trong Postgres để nạp dữ liệu.
-- Password cho tất cả các user là: Password@123

DO $$
DECLARE
    -- Danh sách các thành phố theo yêu cầu
    v_cities TEXT[] := ARRAY['TP.HCM', 'Thành phố Hà Nội', 'Gia Lai', 'Đà Nẵng', 'Huế', 'Ninh Thuận', 'Bình Thuận', 'Đồng Nai'];
    v_city_count INT := array_length(v_cities, 1);
    
    v_branch_id INT;
    i INT;
    v_city_idx INT;
    v_b_code VARCHAR(20);
    
    -- Mảng lưu trữ ID các chi nhánh để quay vòng random
    v_branch_ids INT[];
    
    -- Bcrypt hash cho chuỗi 'Password@123' với cost=10
    v_password_hash VARCHAR(255) := '$2a$10$wh0QEOqJQtRYzRQUypEN3.mMGOlsqF1gchc8hCIdoYVKD9ZXLn7NO';
BEGIN
    -- [Tuỳ chọn] Xóa sạch dữ liệu cũ nếu muốn chạy lại nhiều lần
    -- TRUNCATE wifi_configs, gps_configs, shifts, attendance_logs, users, branches CASCADE;
    
    -- 1. Tạo 99 chi nhánh
    FOR i IN 1..99 LOOP
        -- Trích xuất xoay vòng thành phố từ mảng
        v_city_idx := ((i - 1) % v_city_count) + 1;
        v_b_code := 'CN' || lpad(i::text, 3, '0');
        
        -- Insert Branch (Chi nhánh)
        INSERT INTO branches (code, name, address, city, phone, is_active, created_at, updated_at)
        VALUES (
            v_b_code, 
            'Chi nhánh ' || v_cities[v_city_idx] || ' số ' || i,
            'Địa chỉ tòa nhà HDBank ' || i || ' tại ' || v_cities[v_city_idx], 
            v_cities[v_city_idx], 
            '090' || lpad(i::text, 7, '0'), 
            true, now(), now()
        ) RETURNING id INTO v_branch_id;
        
        -- Thêm ID chi nhánh vừa sinh vào mảng bộ nhớ
        v_branch_ids := array_append(v_branch_ids, v_branch_id);

        -- Insert WiFi Config cho Chi Nhánh
        INSERT INTO wifi_configs (branch_id, ssid, bssid, description, is_active, created_at, updated_at)
        VALUES (
            v_branch_id,
            'HDBank_WIFI_' || v_b_code,
            '00:1A:2B:3C:4D:' || lpad((i % 99)::text, 2, '0'),
            'WiFi Tầng Trệt - ' || v_b_code, true, now(), now()
        );

        -- Insert GPS Config cho Chi Nhánh
        -- Sinh ngẫu nhiên toạ độ tại Việt Nam (Dựa trên mốc trung tâm HCM)
        INSERT INTO gps_configs (branch_id, name, latitude, longitude, radius, description, is_active, created_at, updated_at)
        VALUES (
            v_branch_id,
            'Văn phòng ' || v_b_code,
            10.762622 + (random() * 2.0 - 1.0),   -- Vĩ độ linh tinh tại VN
            106.660172 + (random() * 2.0 - 1.0),  -- Kinh độ linh tinh tại VN
            50.0, 'Bán kính 50m quanh chi nhánh', true, now(), now()
        );

        -- Insert Shift (Ca làm việc: 8:00 - 17:00)
        INSERT INTO shifts (branch_id, name, start_time, end_time, late_after, early_before, work_hours, is_default, is_active, created_at, updated_at)
        VALUES (
            v_branch_id,
            'Ca hành chính',
            '08:00', '17:00', 15, 15, 8.0, true, true, now(), now()
        );
        
        -- Insert Quản lý (Khởi tạo 1 Manager cho mỗi chi nhánh)
        INSERT INTO users (branch_id, employee_code, name, email, phone, password, role, is_active, created_at, updated_at)
        VALUES (
            v_branch_id,
            'MNG' || lpad(i::text, 3, '0'),
            'Quản lý ' || v_b_code,
            'manager_' || v_b_code || '@hdbank.com.vn',
            '091' || lpad(i::text, 7, '0'),
            v_password_hash,
            'manager', true, now(), now()
        );
    END LOOP;

    -- 2. Tạo phần nhân viên còn lại (1000 - 99 Managers = 901 Nhân viên)
    FOR i IN 1..901 LOOP
        -- Chọn ngẫu nhiên 1 Chi nhánh từ mảng 99 chi nhánh
        v_branch_id := v_branch_ids[floor(random() * array_length(v_branch_ids, 1))::int + 1];

        -- Insert Nhân viên (Employee)
        INSERT INTO users (branch_id, employee_code, name, email, phone, password, role, is_active, created_at, updated_at)
        VALUES (
            v_branch_id,
            'EMP' || lpad(i::text, 4, '0'),
            'Nhân viên thứ ' || i,
            'employee_' || i || '@hdbank.com.vn',
            '092' || lpad(i::text, 7, '0'),
            v_password_hash,
            'employee', true, now(), now()
        );
    END LOOP;

    RAISE NOTICE 'Quá trình gieo dữ liệu thành công: 99 Chi nhánh, 99 Managers, 901 Nhân viên bình thường.';
END $$;
