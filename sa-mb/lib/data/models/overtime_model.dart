/// Model yêu cầu tăng ca (OT)
class OvertimeModel {
  final int id;
  final int userId;
  final int branchId;
  final DateTime date;
  final DateTime? actualCheckin;
  final DateTime? actualCheckout;
  final DateTime? calculatedStart;
  final DateTime? calculatedEnd;
  final double totalHours;
  final String status; // pending, approved, rejected
  final int? managerId;
  final DateTime? processedAt;
  final String managerNote;
  final DateTime createdAt;

  // Nested relations
  final Map<String, dynamic>? user;
  final Map<String, dynamic>? branch;
  final Map<String, dynamic>? processedBy;

  OvertimeModel({
    required this.id,
    required this.userId,
    required this.branchId,
    required this.date,
    this.actualCheckin,
    this.actualCheckout,
    this.calculatedStart,
    this.calculatedEnd,
    this.totalHours = 0,
    required this.status,
    this.managerId,
    this.processedAt,
    this.managerNote = '',
    required this.createdAt,
    this.user,
    this.branch,
    this.processedBy,
  });

  factory OvertimeModel.fromJson(Map<String, dynamic> json) {
    return OvertimeModel(
      id: json['id'] as int? ?? 0,
      userId: json['user_id'] as int? ?? 0,
      branchId: json['branch_id'] as int? ?? 0,
      date: DateTime.parse(json['date'] as String).toLocal(),
      actualCheckin: json['actual_checkin'] != null
          ? DateTime.parse(json['actual_checkin'] as String).toLocal()
          : null,
      actualCheckout: json['actual_checkout'] != null
          ? DateTime.parse(json['actual_checkout'] as String).toLocal()
          : null,
      calculatedStart: json['calculated_start'] != null
          ? DateTime.parse(json['calculated_start'] as String).toLocal()
          : null,
      calculatedEnd: json['calculated_end'] != null
          ? DateTime.parse(json['calculated_end'] as String).toLocal()
          : null,
      totalHours: (json['total_hours'] as num?)?.toDouble() ?? 0,
      status: json['status'] as String? ?? 'pending',
      managerId: json['manager_id'] as int?,
      processedAt: json['processed_at'] != null
          ? DateTime.parse(json['processed_at'] as String).toLocal()
          : null,
      managerNote: json['manager_note'] as String? ?? '',
      createdAt: DateTime.parse(
              json['created_at'] as String? ?? DateTime.now().toIso8601String())
          .toLocal(),
      user: json['user'] as Map<String, dynamic>?,
      branch: json['branch'] as Map<String, dynamic>?,
      processedBy: json['processed_by'] as Map<String, dynamic>?,
    );
  }

  bool get isInit => status == 'init';
  bool get isPending => status == 'pending';
  bool get isApproved => status == 'approved';
  bool get isRejected => status == 'rejected';
  bool get isCheckedIn => actualCheckin != null;
  bool get isCheckedOut => actualCheckout != null;

  String get statusDisplay {
    switch (status) {
      case 'init':
        return 'Đang tăng ca';
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

  String get userName => user?['name'] as String? ?? '';
  String get employeeCode => user?['employee_code'] as String? ?? '';
  String get processedByName => processedBy?['name'] as String? ?? '';
}

/// Response check-in OT kèm thông tin bo tròn
class OvertimeCheckInResponse {
  final OvertimeModel overtimeRequest;
  final DateTime? estimatedStart;
  final DateTime? estimatedEnd;
  final String note;

  OvertimeCheckInResponse({
    required this.overtimeRequest,
    this.estimatedStart,
    this.estimatedEnd,
    this.note = '',
  });

  factory OvertimeCheckInResponse.fromJson(Map<String, dynamic> json) {
    return OvertimeCheckInResponse(
      overtimeRequest:
          OvertimeModel.fromJson(json['overtime_request'] as Map<String, dynamic>),
      estimatedStart: json['estimated_start'] != null
          ? DateTime.parse(json['estimated_start'] as String).toLocal()
          : null,
      estimatedEnd: json['estimated_end'] != null
          ? DateTime.parse(json['estimated_end'] as String).toLocal()
          : null,
      note: json['note'] as String? ?? '',
    );
  }
}

/// Response check-out OT kèm thông tin thời gian dự kiến
class OvertimeCheckOutResponse {
  final OvertimeModel overtimeRequest;
  final DateTime? estimatedStart;
  final DateTime? estimatedEnd;
  final double estimatedHours;
  final String note;

  OvertimeCheckOutResponse({
    required this.overtimeRequest,
    this.estimatedStart,
    this.estimatedEnd,
    this.estimatedHours = 0,
    this.note = '',
  });

  factory OvertimeCheckOutResponse.fromJson(Map<String, dynamic> json) {
    return OvertimeCheckOutResponse(
      overtimeRequest:
          OvertimeModel.fromJson(json['overtime_request'] as Map<String, dynamic>),
      estimatedStart: json['estimated_start'] != null
          ? DateTime.parse(json['estimated_start'] as String).toLocal()
          : null,
      estimatedEnd: json['estimated_end'] != null
          ? DateTime.parse(json['estimated_end'] as String).toLocal()
          : null,
      estimatedHours: (json['estimated_hours'] as num?)?.toDouble() ?? 0,
      note: json['note'] as String? ?? '',
    );
  }
}
