-- Gen bulk data: thêm 99 chi nhánh (tổng 100) + ~4995 nhân viên (tổng ~5000)
-- Chạy SAU seed_data.sql (đã có branch id=1, users id=1..5, shift id=1)
-- Password cho tất cả user: Password@123

DO $$
DECLARE
    v_cities TEXT[] := ARRAY['TP.HCM', 'Thành phố Hà Nội', 'Gia Lai', 'Đà Nẵng', 'Huế', 'Ninh Thuận', 'Bình Thuận', 'Đồng Nai'];
    v_provinces TEXT[] := ARRAY['Thành phố Hồ Chí Minh', 'Thành phố Hà Nội', 'Gia Lai', 'Đà Nẵng', 'Thừa Thiên Huế', 'Ninh Thuận', 'Bình Thuận', 'Đồng Nai'];
    v_departments TEXT[] := ARRAY['Công nghệ', 'Phòng Kinh Doanh', 'Phòng Nhân Sự', 'Phòng Tài Chính', 'Phòng Marketing', 'Phòng Hành Chính'];
    v_positions_emp TEXT[] := ARRAY['Nhân viên kinh doanh', 'Chuyên viên', 'Developer', 'Kế toán viên', 'Nhân viên hỗ trợ'];
    v_city_count INT := array_length(v_cities, 1);
    v_dept_count INT := array_length(v_departments, 1);
    v_pos_count INT := array_length(v_positions_emp, 1);

    v_branch_id INT;
    v_branch_ids INT[];
    v_city_idx INT;
    v_b_code VARCHAR(20);
    i INT;
    j INT;

    -- Bcrypt hash cho 'Password@123' cost=10
    v_password_hash VARCHAR(255) := '$2a$10$wh0QEOqJQtRYzRQUypEN3.mMGOlsqF1gchc8hCIdoYVKD9ZXLn7NO';

    -- Đếm user hiện tại (seed_data.sql đã tạo 5 users)
    v_user_seq INT := 5;
    -- Đếm manager
    v_mgr_idx INT := 0;
    -- Số nhân viên cần tạo thêm (5000 tổng - 5 seed - 99 manager = 4896)
    v_remaining_emp INT;
    -- Mảng chứa số nhân viên mỗi chi nhánh mới
    v_emp_per_branch INT[];
    v_total_assigned INT := 0;
    v_rand_count INT;
BEGIN
    -- ================================================================
    -- 1. Tạo 99 chi nhánh mới (id bắt đầu từ 2, tổng 100 với branch 1 từ seed)
    -- ================================================================
    FOR i IN 2..100 LOOP
        v_city_idx := ((i - 1) % v_city_count) + 1;
        v_b_code := 'CN' || lpad(i::text, 3, '0');

        INSERT INTO branches (code, name, address, city, province, phone, email, is_active, created_at, updated_at)
        VALUES (
            v_b_code,
            'Chi nhánh ' || v_cities[v_city_idx] || ' số ' || i,
            'Địa chỉ tòa nhà HDBank ' || i || ' tại ' || v_cities[v_city_idx],
            v_cities[v_city_idx],
            v_provinces[v_city_idx],
            '090' || lpad(i::text, 7, '0'),
            lower(v_b_code) || '@hdbank.com.vn',
            true, now(), now()
        ) RETURNING id INTO v_branch_id;

        v_branch_ids := array_append(v_branch_ids, v_branch_id);

        -- WiFi Config
        INSERT INTO wifi_configs (branch_id, ssid, bssid, description, is_active, created_at, updated_at)
        VALUES (
            v_branch_id,
            'HDBank_WIFI_' || v_b_code,
            '00:1A:2B:3C:4D:' || lpad((i % 99)::text, 2, '0'),
            'WiFi Tầng Trệt - ' || v_b_code,
            true, now(), now()
        );

        -- GPS Config (toạ độ random quanh Việt Nam)
        INSERT INTO gps_configs (branch_id, name, latitude, longitude, radius, description, is_active, created_at, updated_at)
        VALUES (
            v_branch_id,
            'Văn phòng ' || v_b_code,
            10.762622 + (random() * 2.0 - 1.0),
            106.660172 + (random() * 2.0 - 1.0),
            50.0,
            'Bán kính 50m quanh chi nhánh',
            true, now(), now()
        );

        -- Shift (Ca hành chính 8:00-17:00)
        INSERT INTO shifts (branch_id, name, start_time, end_time, late_after, early_before, work_hours, is_default, is_active, created_at, updated_at)
        VALUES (
            v_branch_id,
            'Ca hành chính',
            '08:00', '17:00', 15, 15, 8.0,
            true, true, now(), now()
        );

        -- Manager cho chi nhánh
        v_user_seq := v_user_seq + 1;
        v_mgr_idx := v_mgr_idx + 1;
        INSERT INTO users (employee_code, name, email, phone, password, department, position, branch_id, role, is_active, created_at, updated_at)
        VALUES (
            'MNG' || lpad(v_mgr_idx::text, 3, '0'),
            'Quản lý ' || v_b_code,
            'manager_' || lower(v_b_code) || '@hdbank.com.vn',
            '091' || lpad(v_mgr_idx::text, 7, '0'),
            v_password_hash,
            'Phòng Kinh Doanh',
            'Giám đốc chi nhánh',
            v_branch_id,
            'manager', true, now(), now()
        );
    END LOOP;

    -- ================================================================
    -- 2. Phân bổ random số nhân viên cho mỗi chi nhánh mới (99 CN)
    --    Tổng cần: 5000 - 5 (seed users) - 99 (managers) = 4896
    -- ================================================================
    v_remaining_emp := 4896;

    -- Random số nhân viên cho 98 chi nhánh đầu, chi nhánh cuối nhận phần còn lại
    FOR i IN 1..98 LOOP
        -- Random từ 30 đến 70 nhân viên mỗi chi nhánh
        v_rand_count := 30 + floor(random() * 41)::int;
        -- Đảm bảo không vượt tổng
        IF v_total_assigned + v_rand_count > v_remaining_emp - 1 THEN
            v_rand_count := GREATEST(1, v_remaining_emp - v_total_assigned - (99 - i));
        END IF;
        v_emp_per_branch := array_append(v_emp_per_branch, v_rand_count);
        v_total_assigned := v_total_assigned + v_rand_count;
    END LOOP;
    -- Chi nhánh cuối nhận phần còn lại
    v_emp_per_branch := array_append(v_emp_per_branch, v_remaining_emp - v_total_assigned);

    -- ================================================================
    -- 3. Tạo nhân viên cho từng chi nhánh mới
    -- ================================================================
    FOR i IN 1..99 LOOP
        v_branch_id := v_branch_ids[i];
        FOR j IN 1..v_emp_per_branch[i] LOOP
            v_user_seq := v_user_seq + 1;
            INSERT INTO users (employee_code, name, email, phone, password, department, position, branch_id, role, is_active, created_at, updated_at)
            VALUES (
                'EMP' || lpad(v_user_seq::text, 5, '0'),
                'Nhân viên ' || v_user_seq,
                'employee_' || v_user_seq || '@hdbank.com.vn',
                '092' || lpad(v_user_seq::text, 7, '0'),
                v_password_hash,
                v_departments[floor(random() * v_dept_count)::int + 1],
                v_positions_emp[floor(random() * v_pos_count)::int + 1],
                v_branch_id,
                'employee', true, now(), now()
            );
        END LOOP;
    END LOOP;

    -- Cập nhật tất cả sequences
    PERFORM setval('branches_id_seq', (SELECT MAX(id) FROM branches));
    PERFORM setval('users_id_seq', (SELECT MAX(id) FROM users));
    PERFORM setval('wifi_configs_id_seq', (SELECT MAX(id) FROM wifi_configs));
    PERFORM setval('gps_configs_id_seq', (SELECT MAX(id) FROM gps_configs));
    PERFORM setval('shifts_id_seq', (SELECT MAX(id) FROM shifts));

    RAISE NOTICE 'Bulk data generated: 99 branches (total 100), 99 managers, % employees. Total users: %',
        v_user_seq - 5 - 99, v_user_seq;
END $$;
