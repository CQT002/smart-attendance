import 'package:equatable/equatable.dart';
import '../../../data/models/leave_model.dart';

abstract class LeaveState extends Equatable {
  const LeaveState();

  @override
  List<Object?> get props => [];
}

class LeaveInitial extends LeaveState {}

class LeaveLoading extends LeaveState {}

/// Yêu cầu nghỉ phép được tạo thành công
class LeaveCreateSuccess extends LeaveState {
  final LeaveModel leave;

  const LeaveCreateSuccess(this.leave);

  @override
  List<Object?> get props => [leave];
}

/// Danh sách yêu cầu của employee
class LeaveMyListLoaded extends LeaveState {
  final List<LeaveModel> leaves;

  const LeaveMyListLoaded(this.leaves);

  @override
  List<Object?> get props => [leaves];
}

/// Danh sách yêu cầu cho manager duyệt
class LeaveAdminListLoaded extends LeaveState {
  final List<LeaveModel> leaves;

  const LeaveAdminListLoaded(this.leaves);

  @override
  List<Object?> get props => [leaves];
}

/// Duyệt/từ chối thành công
class LeaveProcessSuccess extends LeaveState {
  final LeaveModel leave;
  final String message;

  const LeaveProcessSuccess(this.leave, this.message);

  @override
  List<Object?> get props => [leave, message];
}

class LeaveBatchApproveSuccess extends LeaveState {
  final int approvedCount;

  const LeaveBatchApproveSuccess(this.approvedCount);

  @override
  List<Object?> get props => [approvedCount];
}

class LeaveFailure extends LeaveState {
  final String message;

  const LeaveFailure(this.message);

  @override
  List<Object?> get props => [message];
}
