import 'package:flutter/material.dart';

class AppColors {
  // Brand colors — tông xanh dương/trắng giống hình mẫu
  static const Color primary = Color(0xFF1565C0);       // Blue 800
  static const Color primaryDark = Color(0xFF0D47A1);   // Blue 900
  static const Color primaryLight = Color(0xFF42A5F5);  // Blue 400
  static const Color secondary = Color(0xFFFF8F00);     // Amber 800
  static const Color secondaryLight = Color(0xFFFFCA28); // Amber 400
  static const Color secondaryDark = Color(0xFFE65100);

  // Status colors
  static const Color success = Color(0xFF43A047);       // Green 600
  static const Color warning = Color(0xFFFB8C00);       // Orange 600
  static const Color error = Color(0xFFE53935);         // Red 600
  static const Color info = Color(0xFF1E88E5);          // Blue 600

  // Calendar status colors
  static const Color calendarPresent = Color(0xFF43A047);   // Xanh lá — đủ công
  static const Color calendarIncomplete = Color(0xFFFB8C00); // Cam — thiếu check-in/out hoặc < 8h
  static const Color calendarAbsent = Color(0xFFE53935);     // Đỏ — chưa chấm công
  static const Color calendarLeave = Color(0xFF1E88E5);      // Xanh dương — nghỉ có phép
  static const Color calendarDayOff = Color(0xFFBDBDBD);     // Xám — cuối tuần/ngày lễ

  // Attendance status
  static const Color statusPresent = Color(0xFF43A047);
  static const Color statusLate = Color(0xFFFB8C00);
  static const Color statusEarlyLeave = Color(0xFFFF7043);
  static const Color statusAbsent = Color(0xFFE53935);
  static const Color statusHalfDay = Color(0xFF7E57C2);

  // Neutral — nền trắng sáng
  static const Color background = Color(0xFFF5F7FA);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color textPrimary = Color(0xFF263238);
  static const Color textSecondary = Color(0xFF78909C);
  static const Color divider = Color(0xFFE0E0E0);

  // Dark theme (kept for compatibility)
  static const Color darkBackground = Color(0xFF121212);
  static const Color darkSurface = Color(0xFF1E1E1E);
  static const Color darkCard = Color(0xFF2C2C2C);

  static Color statusColor(String status) {
    switch (status) {
      case 'present':
        return statusPresent;       // Xanh lá
      case 'late':
      case 'early_leave':
      case 'late_early_leave':
      case 'half_day':
        return statusLate;          // Cam — gom thành "Đi trễ - Về sớm"
      case 'absent':
        return statusAbsent;        // Đỏ
      default:
        return textSecondary;
    }
  }
}
