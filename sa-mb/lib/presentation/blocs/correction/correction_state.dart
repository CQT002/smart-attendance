import 'package:equatable/equatable.dart';
import '../../../data/models/correction_model.dart';

abstract class CorrectionState extends Equatable {
  const CorrectionState();

  @override
  List<Object?> get props => [];
}

class CorrectionInitial extends CorrectionState {}

class CorrectionLoading extends CorrectionState {}

/// Yêu cầu bù công được tạo thành công
class CorrectionCreateSuccess extends CorrectionState {
  final CorrectionModel correction;

  const CorrectionCreateSuccess(this.correction);

  @override
  List<Object?> get props => [correction];
}

/// Danh sách yêu cầu của employee
class CorrectionMyListLoaded extends CorrectionState {
  final List<CorrectionModel> corrections;

  const CorrectionMyListLoaded(this.corrections);

  @override
  List<Object?> get props => [corrections];
}

/// Danh sách yêu cầu cho manager duyệt
class CorrectionAdminListLoaded extends CorrectionState {
  final List<CorrectionModel> corrections;

  const CorrectionAdminListLoaded(this.corrections);

  @override
  List<Object?> get props => [corrections];
}

/// Duyệt/từ chối thành công
class CorrectionProcessSuccess extends CorrectionState {
  final CorrectionModel correction;
  final String message;

  const CorrectionProcessSuccess(this.correction, this.message);

  @override
  List<Object?> get props => [correction, message];
}

/// Duyệt hàng loạt thành công
class CorrectionBatchApproveSuccess extends CorrectionState {
  final int approvedCount;

  const CorrectionBatchApproveSuccess(this.approvedCount);

  @override
  List<Object?> get props => [approvedCount];
}

class CorrectionFailure extends CorrectionState {
  final String message;

  const CorrectionFailure(this.message);

  @override
  List<Object?> get props => [message];
}
