import 'package:flutter/material.dart';

class AppColors {
  // HDBank brand colors (đỏ - vàng - trắng)
  static const Color primary = Color(0xFFCE1126);       // HDBank Red
  static const Color primaryDark = Color(0xFFA30E1F);
  static const Color primaryLight = Color(0xFFE8455A);
  static const Color secondary = Color(0xFFFFB81C);     // HDBank Gold/Yellow
  static const Color secondaryLight = Color(0xFFFFD166);
  static const Color secondaryDark = Color(0xFFE5A000);

  // Status colors
  static const Color success = Color(0xFF4CAF50);
  static const Color warning = Color(0xFFFFA726);
  static const Color error = Color(0xFFD32F2F);
  static const Color info = Color(0xFF42A5F5);

  // Attendance status
  static const Color statusPresent = Color(0xFF4CAF50);
  static const Color statusLate = Color(0xFFFFA726);
  static const Color statusEarlyLeave = Color(0xFFFF7043);
  static const Color statusAbsent = Color(0xFFE53935);
  static const Color statusHalfDay = Color(0xFF7E57C2);

  // Neutral
  static const Color background = Color(0xFFFAF9F7);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color textPrimary = Color(0xFF2D2D2D);
  static const Color textSecondary = Color(0xFF757575);
  static const Color divider = Color(0xFFE0E0E0);

  // Dark theme
  static const Color darkBackground = Color(0xFF121212);
  static const Color darkSurface = Color(0xFF1E1E1E);
  static const Color darkCard = Color(0xFF2C2C2C);

  static Color statusColor(String status) {
    switch (status) {
      case 'present':
        return statusPresent;
      case 'late':
        return statusLate;
      case 'early_leave':
        return statusEarlyLeave;
      case 'absent':
        return statusAbsent;
      case 'half_day':
        return statusHalfDay;
      default:
        return textSecondary;
    }
  }
}
