class AppConstants {
  static const String appName = 'Smart Attendance';
  static const String tokenKey = 'access_token';
  static const String refreshTokenKey = 'refresh_token';
  static const String userKey = 'user_data';

  // Attendance status
  static const String statusPresent = 'present';
  static const String statusLate = 'late';
  static const String statusEarlyLeave = 'early_leave';
  static const String statusLateEarlyLeave = 'late_early_leave';
  static const String statusAbsent = 'absent';
  static const String statusHalfDay = 'half_day';
  static const String statusLeave = 'leave';
  static const String statusHalfDayLeave = 'half_day_leave';

  // Check methods
  static const String methodWifi = 'wifi';
  static const String methodGps = 'gps';
}
