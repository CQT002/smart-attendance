import 'package:equatable/equatable.dart';

abstract class LeaveEvent extends Equatable {
  const LeaveEvent();

  @override
  List<Object?> get props => [];
}

/// Employee: tạo yêu cầu nghỉ phép
class LeaveCreateRequested extends LeaveEvent {
  final String leaveDate;
  final String leaveType;
  final String description;

  const LeaveCreateRequested({
    required this.leaveDate,
    required this.leaveType,
    required this.description,
  });

  @override
  List<Object?> get props => [leaveDate, leaveType, description];
}

/// Employee: load danh sách yêu cầu của mình
class LeaveLoadMyList extends LeaveEvent {
  final String? status;

  const LeaveLoadMyList({this.status});

  @override
  List<Object?> get props => [status];
}

/// Manager: load danh sách yêu cầu cần duyệt
class LeaveLoadAdminList extends LeaveEvent {
  final String? status;

  const LeaveLoadAdminList({this.status});

  @override
  List<Object?> get props => [status];
}

/// Manager: duyệt yêu cầu
class LeaveApproveRequested extends LeaveEvent {
  final int leaveId;
  final String managerNote;

  const LeaveApproveRequested({
    required this.leaveId,
    this.managerNote = '',
  });

  @override
  List<Object?> get props => [leaveId, managerNote];
}

/// Manager: từ chối yêu cầu
class LeaveRejectRequested extends LeaveEvent {
  final int leaveId;
  final String managerNote;

  const LeaveRejectRequested({
    required this.leaveId,
    this.managerNote = '',
  });

  @override
  List<Object?> get props => [leaveId, managerNote];
}

/// Manager: duyệt tất cả yêu cầu đang chờ
class LeaveBatchApproveRequested extends LeaveEvent {}
