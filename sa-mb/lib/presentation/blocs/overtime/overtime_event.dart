import 'package:equatable/equatable.dart';

abstract class OvertimeEvent extends Equatable {
  const OvertimeEvent();

  @override
  List<Object?> get props => [];
}

/// Employee: check-in tăng ca
class OvertimeCheckInRequested extends OvertimeEvent {}

/// Employee: check-out tăng ca
class OvertimeCheckOutRequested extends OvertimeEvent {
  final int? overtimeId;

  const OvertimeCheckOutRequested({this.overtimeId});

  @override
  List<Object?> get props => [overtimeId];
}

/// Employee: load trạng thái OT hôm nay
class OvertimeLoadToday extends OvertimeEvent {}

/// Employee: load lịch sử OT
class OvertimeLoadMyList extends OvertimeEvent {
  final String? status;

  const OvertimeLoadMyList({this.status});

  @override
  List<Object?> get props => [status];
}

/// Manager: load danh sách yêu cầu OT
class OvertimeLoadAdminList extends OvertimeEvent {
  final String? status;

  const OvertimeLoadAdminList({this.status});

  @override
  List<Object?> get props => [status];
}

/// Manager: duyệt yêu cầu OT
class OvertimeApproveRequested extends OvertimeEvent {
  final int overtimeId;
  final String managerNote;

  const OvertimeApproveRequested({
    required this.overtimeId,
    this.managerNote = '',
  });

  @override
  List<Object?> get props => [overtimeId, managerNote];
}

/// Manager: từ chối yêu cầu OT
class OvertimeRejectRequested extends OvertimeEvent {
  final int overtimeId;
  final String managerNote;

  const OvertimeRejectRequested({
    required this.overtimeId,
    this.managerNote = '',
  });

  @override
  List<Object?> get props => [overtimeId, managerNote];
}

/// Manager: duyệt tất cả yêu cầu OT đang chờ
class OvertimeBatchApproveRequested extends OvertimeEvent {}
