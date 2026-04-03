import 'branch_model.dart';

class UserModel {
  final int id;
  final String employeeCode;
  final String name;
  final String email;
  final String phone;
  final String department;
  final String position;
  final String avatarUrl;
  final DateTime? hiredAt;
  final DateTime? lastLoginAt;
  final int? branchId;
  final BranchModel? branch;
  final String role;
  final bool isActive;
  final DateTime createdAt;
  final DateTime updatedAt;

  UserModel({
    required this.id,
    required this.employeeCode,
    required this.name,
    required this.email,
    required this.phone,
    required this.department,
    required this.position,
    this.avatarUrl = '',
    this.hiredAt,
    this.lastLoginAt,
    this.branchId,
    this.branch,
    required this.role,
    required this.isActive,
    required this.createdAt,
    required this.updatedAt,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'] as int? ?? 0,
      employeeCode: json['employee_code'] as String? ?? '',
      name: json['name'] as String? ?? '',
      email: json['email'] as String? ?? '',
      phone: json['phone'] as String? ?? '',
      department: json['department'] as String? ?? '',
      position: json['position'] as String? ?? '',
      avatarUrl: json['avatar_url'] as String? ?? '',
      hiredAt: json['hired_at'] != null ? DateTime.parse(json['hired_at'] as String).toLocal() : null,
      lastLoginAt: json['last_login_at'] != null ? DateTime.parse(json['last_login_at'] as String).toLocal() : null,
      branchId: json['branch_id'] as int?,
      branch: json['branch'] != null ? BranchModel.fromJson(json['branch'] as Map<String, dynamic>) : null,
      role: json['role'] as String? ?? 'employee',
      isActive: json['is_active'] as bool? ?? true,
      createdAt: DateTime.parse(json['created_at'] as String? ?? DateTime.now().toIso8601String()).toLocal(),
      updatedAt: DateTime.parse(json['updated_at'] as String? ?? DateTime.now().toIso8601String()).toLocal(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'employee_code': employeeCode,
      'name': name,
      'email': email,
      'phone': phone,
      'department': department,
      'position': position,
      'avatar_url': avatarUrl,
      'hired_at': hiredAt?.toIso8601String(),
      'last_login_at': lastLoginAt?.toIso8601String(),
      'branch_id': branchId,
      'role': role,
      'is_active': isActive,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  bool get isEmployee => role == 'employee';
  bool get isManager => role == 'manager';
  bool get isAdmin => role == 'admin';
}
