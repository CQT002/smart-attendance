import 'package:flutter/material.dart';

class AppColors {
  // HDBank brand colors
  static const Color primary = Color(0xFFE31837);       // HDBank Red
  static const Color primaryDark = Color(0xFFB71430);
  static const Color primaryLight = Color(0xFFFF5252);
  static const Color secondary = Color(0xFF1B5E20);     // HDBank Green
  static const Color secondaryLight = Color(0xFF4CAF50);

  // Status colors
  static const Color success = Color(0xFF4CAF50);
  static const Color warning = Color(0xFFFFA726);
  static const Color error = Color(0xFFE53935);
  static const Color info = Color(0xFF42A5F5);

  // Attendance status
  static const Color statusPresent = Color(0xFF4CAF50);
  static const Color statusLate = Color(0xFFFFA726);
  static const Color statusEarlyLeave = Color(0xFFFF7043);
  static const Color statusAbsent = Color(0xFFE53935);
  static const Color statusHalfDay = Color(0xFF7E57C2);

  // Neutral
  static const Color background = Color(0xFFF5F5F5);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color textPrimary = Color(0xFF212121);
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
