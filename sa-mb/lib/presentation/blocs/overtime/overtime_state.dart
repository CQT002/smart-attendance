import 'package:equatable/equatable.dart';
import '../../../data/models/overtime_model.dart';

abstract class OvertimeState extends Equatable {
  const OvertimeState();

  @override
  List<Object?> get props => [];
}

class OvertimeInitial extends OvertimeState {}

class OvertimeLoading extends OvertimeState {}

/// Check-in OT thành công
class OvertimeCheckInSuccess extends OvertimeState {
  final OvertimeCheckInResponse response;

  const OvertimeCheckInSuccess(this.response);

  @override
  List<Object?> get props => [response];
}

/// Check-out OT thành công
class OvertimeCheckOutSuccess extends OvertimeState {
  final OvertimeCheckOutResponse response;

  const OvertimeCheckOutSuccess(this.response);

  @override
  List<Object?> get props => [response];
}

/// Trạng thái OT hôm nay
class OvertimeTodayLoaded extends OvertimeState {
  final OvertimeModel? overtime;

  const OvertimeTodayLoaded(this.overtime);

  @override
  List<Object?> get props => [overtime];
}

/// Danh sách OT của employee
class OvertimeMyListLoaded extends OvertimeState {
  final List<OvertimeModel> overtimes;

  const OvertimeMyListLoaded(this.overtimes);

  @override
  List<Object?> get props => [overtimes];
}

/// Danh sách OT cho manager duyệt
class OvertimeAdminListLoaded extends OvertimeState {
  final List<OvertimeModel> overtimes;

  const OvertimeAdminListLoaded(this.overtimes);

  @override
  List<Object?> get props => [overtimes];
}

/// Duyệt/từ chối thành công
class OvertimeProcessSuccess extends OvertimeState {
  final OvertimeModel overtime;
  final String message;

  const OvertimeProcessSuccess(this.overtime, this.message);

  @override
  List<Object?> get props => [overtime, message];
}

class OvertimeBatchApproveSuccess extends OvertimeState {
  final int approvedCount;

  const OvertimeBatchApproveSuccess(this.approvedCount);

  @override
  List<Object?> get props => [approvedCount];
}

class OvertimeFailure extends OvertimeState {
  final String message;

  const OvertimeFailure(this.message);

  @override
  List<Object?> get props => [message];
}
