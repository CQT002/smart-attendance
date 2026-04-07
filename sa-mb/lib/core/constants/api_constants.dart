class ApiConstants {
  static const String baseUrl = 'http://localhost:8080';
  static const String apiPrefix = '/api/v1';

  // Auth
  static const String login = '$apiPrefix/auth/login';
  static const String me = '$apiPrefix/auth/me';
  static const String changePassword = '$apiPrefix/auth/change-password';

  // Attendance
  static const String checkIn = '$apiPrefix/attendance/check-in';
  static const String checkOut = '$apiPrefix/attendance/check-out';
  static const String todayAttendance = '$apiPrefix/attendance/today';
  static const String attendanceHistory = '$apiPrefix/attendance/history';

  // Corrections (employee)
  static const String corrections = '$apiPrefix/attendance/corrections';

  // Corrections (manager/admin)
  static const String adminCorrections = '$apiPrefix/admin/corrections';
}
