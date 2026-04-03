-- Cập nhật dữ liệu cho bảng users: bổ sung Chức vụ và Phòng ban
UPDATE users
SET department = 'Phòng Kinh Doanh',
    position = CASE 
        WHEN role = 'manager' THEN 'Giám đốc chi nhánh'
        WHEN role = 'admin' THEN 'Giám đốc hệ thống'
        ELSE 'Nhân viên kinh doanh'
    END
WHERE department IS NULL OR position IS NULL;

-- Cập nhật dữ liệu cho bảng branches: bổ sung email
UPDATE branches
SET email = lower(code) || '@hdbank.com.vn'
WHERE email IS NULL;
