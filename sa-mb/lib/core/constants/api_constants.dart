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

  // Leaves (employee)
  static const String leaves = '$apiPrefix/attendance/leaves';

  // Leaves (manager/admin)
  static const String adminLeaves = '$apiPrefix/admin/leaves';

  // Batch approve
  static const String batchApproveCorrections = '$apiPrefix/admin/corrections/batch-approve';
  static const String batchApproveLeaves = '$apiPrefix/admin/leaves/batch-approve';

  // Overtime (employee)
  static const String overtime = '$apiPrefix/attendance/overtime';
  static const String overtimeCheckIn = '$apiPrefix/attendance/overtime/check-in';
  static const String overtimeCheckOut = '$apiPrefix/attendance/overtime/check-out';
  static const String overtimeToday = '$apiPrefix/attendance/overtime/today';

  // Overtime (manager/admin)
  static const String adminOvertime = '$apiPrefix/admin/overtime';
  static const String batchApproveOvertime = '$apiPrefix/admin/overtime/batch-approve';

  // Unified approvals (manager/admin)
  static const String approvals = '$apiPrefix/admin/approvals';
  static const String pendingApprovals = '$apiPrefix/admin/approvals/pending';
}
