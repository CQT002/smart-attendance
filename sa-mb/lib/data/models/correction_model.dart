import 'attendance_model.dart';
import 'user_model.dart';

class CorrectionModel {
  final int id;
  final String correctionType; // "attendance" | "overtime"
  final int userId;
  final int branchId;
  final int? attendanceLogId;
  final int? overtimeRequestId;
  final String originalStatus;
  final int creditCount;
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
  final AttendanceModel? attendanceLog;

  CorrectionModel({
    required this.id,
    this.correctionType = 'attendance',
    required this.userId,
    required this.branchId,
    this.attendanceLogId,
    this.overtimeRequestId,
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
      correctionType: json['correction_type'] as String? ?? 'attendance',
      userId: json['user_id'] as int? ?? 0,
      branchId: json['branch_id'] as int? ?? 0,
      attendanceLogId: json['attendance_log_id'] as int?,
      overtimeRequestId: json['overtime_request_id'] as int?,
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
  bool get isOvertime => correctionType == 'overtime';

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

  String get correctionTypeDisplay =>
      isOvertime ? 'Bổ sung công tăng ca' : 'Bổ sung công ca chính thức';

  String get originalStatusDisplay {
    switch (originalStatus) {
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
      case 'missing_checkin':
        return 'Thiếu check-in OT';
      case 'missing_checkout':
        return 'Thiếu check-out OT';
      default:
        return originalStatus;
    }
  }
}
