import 'package:equatable/equatable.dart';

abstract class CorrectionEvent extends Equatable {
  const CorrectionEvent();

  @override
  List<Object?> get props => [];
}

/// Employee: tạo yêu cầu bù công
class CorrectionCreateRequested extends CorrectionEvent {
  final int attendanceLogId;
  final String description;

  const CorrectionCreateRequested({
    required this.attendanceLogId,
    required this.description,
  });

  @override
  List<Object?> get props => [attendanceLogId, description];
}

/// Employee: load danh sách yêu cầu của mình
class CorrectionLoadMyList extends CorrectionEvent {
  final String? status;

  const CorrectionLoadMyList({this.status});

  @override
  List<Object?> get props => [status];
}

/// Manager: load danh sách yêu cầu cần duyệt
class CorrectionLoadAdminList extends CorrectionEvent {
  final String? status;

  const CorrectionLoadAdminList({this.status});

  @override
  List<Object?> get props => [status];
}

/// Manager: duyệt yêu cầu
class CorrectionApproveRequested extends CorrectionEvent {
  final int correctionId;
  final String managerNote;

  const CorrectionApproveRequested({
    required this.correctionId,
    this.managerNote = '',
  });

  @override
  List<Object?> get props => [correctionId, managerNote];
}

/// Manager: từ chối yêu cầu
class CorrectionRejectRequested extends CorrectionEvent {
  final int correctionId;
  final String managerNote;

  const CorrectionRejectRequested({
    required this.correctionId,
    this.managerNote = '',
  });

  @override
  List<Object?> get props => [correctionId, managerNote];
}

/// Manager: duyệt tất cả yêu cầu đang chờ
class CorrectionBatchApproveRequested extends CorrectionEvent {}
