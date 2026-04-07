class AttendanceModel {
  final int id;
  final int userId;
  final int branchId;
  final int? shiftId;
  final DateTime date;

  // Check-in
  final DateTime? checkInTime;
  final double? checkInLat;
  final double? checkInLng;
  final String? checkInMethod;
  final String checkInSsid;
  final String checkInBssid;

  // Check-out
  final DateTime? checkOutTime;
  final double? checkOutLat;
  final double? checkOutLng;
  final String? checkOutMethod;
  final String checkOutSsid;
  final String checkOutBssid;

  // Device
  final String deviceId;
  final String deviceModel;
  final String ipAddress;
  final String appVersion;

  // Anti-fraud
  final bool isFakeGps;
  final bool isVpn;
  final String fraudNote;

  // Calculated
  final String status; // present | late | early_leave | absent | half_day
  final double workHours;
  final double overtime;
  final String note;
  final DateTime createdAt;
  final DateTime updatedAt;

  AttendanceModel({
    required this.id,
    required this.userId,
    required this.branchId,
    this.shiftId,
    required this.date,
    this.checkInTime,
    this.checkInLat,
    this.checkInLng,
    this.checkInMethod,
    this.checkInSsid = '',
    this.checkInBssid = '',
    this.checkOutTime,
    this.checkOutLat,
    this.checkOutLng,
    this.checkOutMethod,
    this.checkOutSsid = '',
    this.checkOutBssid = '',
    this.deviceId = '',
    this.deviceModel = '',
    this.ipAddress = '',
    this.appVersion = '',
    this.isFakeGps = false,
    this.isVpn = false,
    this.fraudNote = '',
    required this.status,
    this.workHours = 0,
    this.overtime = 0,
    this.note = '',
    required this.createdAt,
    required this.updatedAt,
  });

  factory AttendanceModel.fromJson(Map<String, dynamic> json) {
    return AttendanceModel(
      id: json['id'] as int? ?? 0,
      userId: json['user_id'] as int? ?? 0,
      branchId: json['branch_id'] as int? ?? 0,
      shiftId: json['shift_id'] as int?,
      date: DateTime.parse(json['date'] as String? ?? DateTime.now().toIso8601String()).toLocal(),
      checkInTime: json['check_in_time'] != null ? DateTime.parse(json['check_in_time'] as String).toLocal() : null,
      checkInLat: (json['check_in_lat'] as num?)?.toDouble(),
      checkInLng: (json['check_in_lng'] as num?)?.toDouble(),
      checkInMethod: json['check_in_method'] as String?,
      checkInSsid: json['check_in_ssid'] as String? ?? '',
      checkInBssid: json['check_in_bssid'] as String? ?? '',
      checkOutTime: json['check_out_time'] != null ? DateTime.parse(json['check_out_time'] as String).toLocal() : null,
      checkOutLat: (json['check_out_lat'] as num?)?.toDouble(),
      checkOutLng: (json['check_out_lng'] as num?)?.toDouble(),
      checkOutMethod: json['check_out_method'] as String?,
      checkOutSsid: json['check_out_ssid'] as String? ?? '',
      checkOutBssid: json['check_out_bssid'] as String? ?? '',
      deviceId: json['device_id'] as String? ?? '',
      deviceModel: json['device_model'] as String? ?? '',
      ipAddress: json['ip_address'] as String? ?? '',
      appVersion: json['app_version'] as String? ?? '',
      isFakeGps: json['is_fake_gps'] as bool? ?? false,
      isVpn: json['is_vpn'] as bool? ?? false,
      fraudNote: json['fraud_note'] as String? ?? '',
      status: json['status'] as String? ?? 'absent',
      workHours: (json['work_hours'] as num?)?.toDouble() ?? 0,
      overtime: (json['overtime'] as num?)?.toDouble() ?? 0,
      note: json['note'] as String? ?? '',
      createdAt: DateTime.parse(json['created_at'] as String? ?? DateTime.now().toIso8601String()).toLocal(),
      updatedAt: DateTime.parse(json['updated_at'] as String? ?? DateTime.now().toIso8601String()).toLocal(),
    );
  }

  bool get hasCheckedIn => checkInTime != null;
  bool get hasCheckedOut => checkOutTime != null;
  bool get isComplete => hasCheckedIn && hasCheckedOut;

  String get statusDisplay {
    switch (status) {
      case 'present':
        return 'Đúng giờ';
      case 'late':
      case 'early_leave':
      case 'late_early_leave':
      case 'half_day':
        return 'Đi trễ - Về sớm';
      case 'absent':
        return 'Vắng';
      case 'leave':
        return 'Nghỉ phép';
      case 'half_day_leave':
        return 'Nghỉ phép nửa ngày';
      default:
        return status;
    }
  }
}
