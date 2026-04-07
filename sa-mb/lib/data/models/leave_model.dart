import 'user_model.dart';

class LeaveModel {
  final int id;
  final int userId;
  final int branchId;
  final String leaveDate;
  final String leaveType; // full_day | half_day_morning | half_day_afternoon
  final String timeFrom;
  final String timeTo;
  final String originalStatus;
  final String description;
  final String status; // pending | approved | rejected
  final int? processedById;
  final DateTime? processedAt;
  final String managerNote;
  final DateTime createdAt;
  final DateTime updatedAt;

  // Relations
  final UserModel? user;
  final UserModel? processedBy;

  LeaveModel({
    required this.id,
    required this.userId,
    required this.branchId,
    required this.leaveDate,
    required this.leaveType,
    required this.timeFrom,
    required this.timeTo,
    this.originalStatus = '',
    required this.description,
    required this.status,
    this.processedById,
    this.processedAt,
    this.managerNote = '',
    required this.createdAt,
    required this.updatedAt,
    this.user,
    this.processedBy,
  });

  factory LeaveModel.fromJson(Map<String, dynamic> json) {
    return LeaveModel(
      id: json['id'] as int? ?? 0,
      userId: json['user_id'] as int? ?? 0,
      branchId: json['branch_id'] as int? ?? 0,
      leaveDate: json['leave_date'] as String? ?? '',
      leaveType: json['leave_type'] as String? ?? 'full_day',
      timeFrom: json['time_from'] as String? ?? '',
      timeTo: json['time_to'] as String? ?? '',
      originalStatus: json['original_status'] as String? ?? '',
      description: json['description'] as String? ?? '',
      status: json['status'] as String? ?? 'pending',
      processedById: json['processed_by_id'] as int?,
      processedAt: json['processed_at'] != null
          ? DateTime.parse(json['processed_at'] as String).toLocal()
          : null,
      managerNote: json['manager_note'] as String? ?? '',
      createdAt: DateTime.parse(
              json['created_at'] as String? ?? DateTime.now().toIso8601String())
          .toLocal(),
      updatedAt: DateTime.parse(
              json['updated_at'] as String? ?? DateTime.now().toIso8601String())
          .toLocal(),
      user: json['user'] != null && json['user'] is Map<String, dynamic>
          ? UserModel.fromJson(json['user'] as Map<String, dynamic>)
          : null,
      processedBy: json['processed_by'] != null &&
              json['processed_by'] is Map<String, dynamic>
          ? UserModel.fromJson(json['processed_by'] as Map<String, dynamic>)
          : null,
    );
  }

  bool get isPending => status == 'pending';
  bool get isApproved => status == 'approved';
  bool get isRejected => status == 'rejected';

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
