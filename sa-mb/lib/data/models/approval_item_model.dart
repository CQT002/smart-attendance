/// Model tổng hợp duyệt chấm công — map từ API GET /admin/approvals
class ApprovalItemModel {
  final int id;
  final String type; // "correction" hoặc "leave"
  final int userId;
  final String userName;
  final String employeeCode;
  final String department;
  final int branchId;
  final String date;
  final String description;
  final String detail;
  final String status; // pending, approved, rejected
  final DateTime createdAt;

  // Audit fields
  final String processedByName;
  final DateTime? processedAt;
  final String managerNote;

  // Correction-specific
  final String? checkInTime;
  final String? checkOutTime;

  // Leave-specific
  final String leaveType;
  final String timeFrom;
  final String timeTo;

  // Overtime-specific
  final String? actualCheckin;
  final String? actualCheckout;
  final String? calculatedStart;
  final String? calculatedEnd;
  final double? totalHours;

  ApprovalItemModel({
    required this.id,
    required this.type,
    required this.userId,
    required this.userName,
    required this.employeeCode,
    required this.department,
    required this.branchId,
    required this.date,
    required this.description,
    required this.detail,
    required this.status,
    required this.createdAt,
    this.processedByName = '',
    this.processedAt,
    this.managerNote = '',
    this.checkInTime,
    this.checkOutTime,
    this.leaveType = '',
    this.timeFrom = '',
    this.timeTo = '',
    this.actualCheckin,
    this.actualCheckout,
    this.calculatedStart,
    this.calculatedEnd,
    this.totalHours,
  });

  factory ApprovalItemModel.fromJson(Map<String, dynamic> json) {
    return ApprovalItemModel(
      id: json['id'] as int? ?? 0,
      type: json['type'] as String? ?? '',
      userId: json['user_id'] as int? ?? 0,
      userName: json['user_name'] as String? ?? '',
      employeeCode: json['employee_code'] as String? ?? '',
      department: json['department'] as String? ?? '',
      branchId: json['branch_id'] as int? ?? 0,
      date: json['date'] as String? ?? '',
      description: json['description'] as String? ?? '',
      detail: json['detail'] as String? ?? '',
      status: json['status'] as String? ?? 'pending',
      createdAt: DateTime.parse(
              json['created_at'] as String? ?? DateTime.now().toIso8601String())
          .toLocal(),
      processedByName: json['processed_by_name'] as String? ?? '',
      processedAt: json['processed_at'] != null
          ? DateTime.parse(json['processed_at'] as String).toLocal()
          : null,
      managerNote: json['manager_note'] as String? ?? '',
      checkInTime: json['check_in_time'] as String?,
      checkOutTime: json['check_out_time'] as String?,
      leaveType: json['leave_type'] as String? ?? '',
      timeFrom: json['time_from'] as String? ?? '',
      timeTo: json['time_to'] as String? ?? '',
      actualCheckin: json['actual_checkin'] as String?,
      actualCheckout: json['actual_checkout'] as String?,
      calculatedStart: json['calculated_start'] as String?,
      calculatedEnd: json['calculated_end'] as String?,
      totalHours: (json['total_hours'] as num?)?.toDouble(),
    );
  }

  bool get isPending => status == 'pending';
  bool get isCorrection => type == 'correction';
  bool get isLeave => type == 'leave';
  bool get isOvertime => type == 'overtime';

  String get statusDisplay {
    switch (status) {
      case 'pending':
        return 'Chờ duyệt';
      case 'approved':
        return 'Đã duyệt';
      case 'rejected':
        return 'Từ chối';
      default:
        return status;
    }
  }

  String get typeDisplay {
    if (isCorrection) return 'Bổ sung công';
    if (isLeave) return 'Nghỉ phép';
    if (isOvertime) return 'Tăng ca';
    return type;
  }

  String get originalStatusDisplay {
    switch (detail) {
      case 'present':
        return 'Đúng giờ';
      case 'late':
        return 'Đi trễ';
      case 'early_leave':
        return 'Về sớm';
      case 'late_early_leave':
        return 'Đi trễ - Về sớm';
      case 'absent':
        return 'Vắng mặt';
      case 'half_day':
        return 'Nửa ngày';
      case 'leave':
        return 'Nghỉ phép';
      case 'half_day_leave':
        return 'Nghỉ phép nửa ngày';
      case 'overtime':
        return 'Tăng ca';
      case 'missing_checkin':
        return 'Thiếu check-in OT';
      case 'missing_checkout':
        return 'Thiếu check-out OT';
      default:
        return detail;
    }
  }

  String get leaveTypeDisplay {
    switch (leaveType) {
      case 'full_day':
        return 'Cả ngày';
      case 'half_day_morning':
        return 'Buổi sáng';
      case 'half_day_afternoon':
        return 'Buổi chiều';
      default:
        return leaveType;
    }
  }

  String get timeRangeDisplay => '$timeFrom - $timeTo';
}
