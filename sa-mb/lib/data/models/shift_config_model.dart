/// Cấu hình ca làm việc mặc định của chi nhánh
class ShiftConfigModel {
  final int id;
  final int branchId;
  final String name;
  final String startTime;
  final String endTime;
  final int lateAfter;
  final int earlyBefore;
  final double workHours;
  final String morningEnd;
  final String afternoonStart;
  final int regularEndDay; // 0=CN, 1=T2, ..., 6=T7
  final String regularEndTime; // HH:MM
  final int otMinCheckinHour;
  final int otStartHour;
  final int otEndHour;

  ShiftConfigModel({
    required this.id,
    required this.branchId,
    required this.name,
    required this.startTime,
    required this.endTime,
    this.lateAfter = 15,
    this.earlyBefore = 15,
    this.workHours = 8,
    this.morningEnd = '12:00',
    this.afternoonStart = '13:00',
    this.regularEndDay = 6,
    this.regularEndTime = '12:00',
    this.otMinCheckinHour = 17,
    this.otStartHour = 18,
    this.otEndHour = 22,
  });

  factory ShiftConfigModel.fromJson(Map<String, dynamic> json) {
    return ShiftConfigModel(
      id: json['id'] as int? ?? 0,
      branchId: json['branch_id'] as int? ?? 0,
      name: json['name'] as String? ?? '',
      startTime: json['start_time'] as String? ?? '08:00',
      endTime: json['end_time'] as String? ?? '17:00',
      lateAfter: json['late_after'] as int? ?? 15,
      earlyBefore: json['early_before'] as int? ?? 15,
      workHours: (json['work_hours'] as num?)?.toDouble() ?? 8,
      morningEnd: json['morning_end'] as String? ?? '12:00',
      afternoonStart: json['afternoon_start'] as String? ?? '13:00',
      regularEndDay: json['regular_end_day'] as int? ?? 6,
      regularEndTime: json['regular_end_time'] as String? ?? '12:00',
      otMinCheckinHour: json['ot_min_checkin_hour'] as int? ?? 17,
      otStartHour: json['ot_start_hour'] as int? ?? 18,
      otEndHour: json['ot_end_hour'] as int? ?? 22,
    );
  }

  /// Kiểm tra thời điểm [now] có nằm trong khung giờ làm việc chính thức không
  bool isWithinRegularWindow(DateTime now) {
    final weekday = now.weekday; // Dart: Monday=1, ..., Sunday=7

    // Convert Dart weekday (Mon=1..Sun=7) → Go encoding (Sun=0..Sat=6)
    final goWeekday = weekday == 7 ? 0 : weekday;

    // Edge case: regularEndDay=0 (CN) → cả tuần là chính thức
    if (regularEndDay == 0) {
      if (goWeekday != 0) return true;
      return !_isAfterTime(now, regularEndTime);
    }

    // Chủ nhật luôn ngoài khung khi regularEndDay != 0
    if (goWeekday == 0) return false;

    if (goWeekday < regularEndDay) return true;
    if (goWeekday > regularEndDay) return false;

    // goWeekday == regularEndDay → check giờ
    return !_isAfterTime(now, regularEndTime);
  }

  bool _isAfterTime(DateTime now, String timeStr) {
    final parts = timeStr.split(':');
    if (parts.length != 2) return false;
    final h = int.tryParse(parts[0]) ?? 0;
    final m = int.tryParse(parts[1]) ?? 0;
    final limit = DateTime(now.year, now.month, now.day, h, m);
    return now.isAfter(limit);
  }

  /// Parse giờ từ string HH:MM → (hour, minute)
  static (int, int) parseTime(String t) {
    final parts = t.split(':');
    if (parts.length != 2) return (8, 0);
    return (int.tryParse(parts[0]) ?? 8, int.tryParse(parts[1]) ?? 0);
  }

  /// Default fallback khi chưa fetch được
  static ShiftConfigModel defaultConfig() {
    return ShiftConfigModel(
      id: 0,
      branchId: 0,
      name: 'Mặc định',
      startTime: '08:00',
      endTime: '17:00',
    );
  }
}
