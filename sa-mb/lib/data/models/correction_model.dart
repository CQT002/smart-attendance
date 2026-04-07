import 'attendance_model.dart';
import 'user_model.dart';

class CorrectionModel {
  final int id;
  final int userId;
  final int branchId;
  final int attendanceLogId;
  final String originalStatus;
  final int creditCount; // late/early_leave=1, late_early_leave=2
  final String description;
  final String status; // pending | approved | rejected
  final int? processedById;
  final DateTime? processedAt;
  final String managerNote;
  final DateTime createdAt;
  final DateTime updatedAt;

  // Relations (may be null depending on API preload)
  final UserModel? user;
  final UserModel? processedBy;
  final AttendanceModel? attendanceLog;

  CorrectionModel({
    required this.id,
    required this.userId,
    required this.branchId,
    required this.attendanceLogId,
    required this.originalStatus,
    this.creditCount = 1,
    required this.description,
    required this.status,
    this.processedById,
    this.processedAt,
    this.managerNote = '',
    required this.createdAt,
    required this.updatedAt,
    this.user,
    this.processedBy,
    this.attendanceLog,
  });

  factory CorrectionModel.fromJson(Map<String, dynamic> json) {
    return CorrectionModel(
      id: json['id'] as int? ?? 0,
      userId: json['user_id'] as int? ?? 0,
      branchId: json['branch_id'] as int? ?? 0,
      attendanceLogId: json['attendance_log_id'] as int? ?? 0,
      originalStatus: json['original_status'] as String? ?? '',
      creditCount: json['credit_count'] as int? ?? 1,
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
      attendanceLog: json['attendance_log'] != null &&
              json['attendance_log'] is Map<String, dynamic>
          ? AttendanceModel.fromJson(
              json['attendance_log'] as Map<String, dynamic>)
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

  String get originalStatusDisplay {
    switch (originalStatus) {
      case 'late':
        return 'Đi trễ';
      case 'early_leave':
        return 'Về sớm';
      case 'late_early_leave':
        return 'Đi trễ - Về sớm';
      default:
        return originalStatus;
    }
  }
}
