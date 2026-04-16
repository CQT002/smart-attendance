/// HolidayModel — bản dữ liệu ngày lễ đồng bộ với backend `entity.Holiday`.
class HolidayModel {
  final int id;
  final String name;
  final DateTime date; // date-only
  final int year;
  final double coefficient; // 2.0, 3.0, 4.0...
  final String type; // national | company
  final bool isCompensated;
  final DateTime? compensateFor;
  final String description;

  HolidayModel({
    required this.id,
    required this.name,
    required this.date,
    required this.year,
    required this.coefficient,
    required this.type,
    this.isCompensated = false,
    this.compensateFor,
    this.description = '',
  });

  factory HolidayModel.fromJson(Map<String, dynamic> json) {
    return HolidayModel(
      id: json['id'] as int? ?? 0,
      name: json['name'] as String? ?? '',
      date: DateTime.parse(json['date'] as String? ?? DateTime.now().toIso8601String()).toLocal(),
      year: json['year'] as int? ?? 0,
      coefficient: (json['coefficient'] as num?)?.toDouble() ?? 1.0,
      type: json['type'] as String? ?? 'national',
      isCompensated: json['is_compensated'] as bool? ?? false,
      compensateFor: json['compensate_for'] != null
          ? DateTime.parse(json['compensate_for'] as String).toLocal()
          : null,
      description: json['description'] as String? ?? '',
    );
  }

  /// Hiển thị hệ số dạng "x3.0"
  String get coefficientDisplay => 'x${coefficient.toStringAsFixed(coefficient.truncateToDouble() == coefficient ? 1 : 2)}';

  /// Hiển thị loại lễ
  String get typeDisplay {
    switch (type) {
      case 'national':
        return 'Lễ quốc gia';
      case 'company':
        return 'Lễ công ty';
      default:
        return type;
    }
  }

  /// Key YYYY-MM-DD để dùng map lookup
  String get dateKey =>
      '${date.year.toString().padLeft(4, '0')}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
}

/// DailyAttendance — item per-day trong summary response.
class DailyAttendanceModel {
  final String date; // YYYY-MM-DD
  final int weekday;
  final String status;
  final double workHours;
  final DateTime? checkInTime;
  final DateTime? checkOutTime;
  final bool isHoliday;
  final int? holidayId;
  final String holidayName;
  final String holidayType;
  final double coefficient;
  final double effectiveMultiplier;
  final bool isCompensated;

  DailyAttendanceModel({
    required this.date,
    required this.weekday,
    required this.status,
    this.workHours = 0,
    this.checkInTime,
    this.checkOutTime,
    this.isHoliday = false,
    this.holidayId,
    this.holidayName = '',
    this.holidayType = '',
    this.coefficient = 0,
    this.effectiveMultiplier = 0,
    this.isCompensated = false,
  });

  factory DailyAttendanceModel.fromJson(Map<String, dynamic> json) {
    return DailyAttendanceModel(
      date: json['date'] as String? ?? '',
      weekday: json['weekday'] as int? ?? 0,
      status: json['status'] as String? ?? '',
      workHours: (json['work_hours'] as num?)?.toDouble() ?? 0,
      checkInTime: json['check_in_time'] != null
          ? DateTime.parse(json['check_in_time'] as String).toLocal()
          : null,
      checkOutTime: json['check_out_time'] != null
          ? DateTime.parse(json['check_out_time'] as String).toLocal()
          : null,
      isHoliday: json['is_holiday'] as bool? ?? false,
      holidayId: json['holiday_id'] as int?,
      holidayName: json['holiday_name'] as String? ?? '',
      holidayType: json['holiday_type'] as String? ?? '',
      coefficient: (json['coefficient'] as num?)?.toDouble() ?? 0,
      effectiveMultiplier: (json['effective_multiplier'] as num?)?.toDouble() ?? 0,
      isCompensated: json['is_compensated'] as bool? ?? false,
    );
  }
}

/// AttendanceSummaryModel — response từ GET /attendance/summary.
class AttendanceSummaryModel {
  final int userId;
  final String dateFrom;
  final String dateTo;
  final int totalWorkDays;
  final int regularWorkDays;
  final int holidayWorkDays;
  final int paidHolidayDays;
  final int leaveDays;
  final int absentDays;
  final double totalWorkHours;
  final double totalSalaryMultiplier;
  final List<DailyAttendanceModel> days;

  AttendanceSummaryModel({
    required this.userId,
    required this.dateFrom,
    required this.dateTo,
    this.totalWorkDays = 0,
    this.regularWorkDays = 0,
    this.holidayWorkDays = 0,
    this.paidHolidayDays = 0,
    this.leaveDays = 0,
    this.absentDays = 0,
    this.totalWorkHours = 0,
    this.totalSalaryMultiplier = 0,
    this.days = const [],
  });

  factory AttendanceSummaryModel.fromJson(Map<String, dynamic> json) {
    final daysJson = json['days'] as List<dynamic>? ?? [];
    return AttendanceSummaryModel(
      userId: json['user_id'] as int? ?? 0,
      dateFrom: json['date_from'] as String? ?? '',
      dateTo: json['date_to'] as String? ?? '',
      totalWorkDays: json['total_work_days'] as int? ?? 0,
      regularWorkDays: json['regular_work_days'] as int? ?? 0,
      holidayWorkDays: json['holiday_work_days'] as int? ?? 0,
      paidHolidayDays: json['paid_holiday_days'] as int? ?? 0,
      leaveDays: json['leave_days'] as int? ?? 0,
      absentDays: json['absent_days'] as int? ?? 0,
      totalWorkHours: (json['total_work_hours'] as num?)?.toDouble() ?? 0,
      totalSalaryMultiplier: (json['total_salary_multiplier'] as num?)?.toDouble() ?? 0,
      days: daysJson
          .map((e) => DailyAttendanceModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
