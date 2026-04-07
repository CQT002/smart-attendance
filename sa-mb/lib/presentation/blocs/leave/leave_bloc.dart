import 'package:dio/dio.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../domain/repositories/leave_repository.dart';
import 'leave_event.dart';
import 'leave_state.dart';

class LeaveBloc extends Bloc<LeaveEvent, LeaveState> {
  final LeaveRepository _leaveRepository;

  LeaveBloc({required LeaveRepository leaveRepository})
      : _leaveRepository = leaveRepository,
        super(LeaveInitial()) {
    on<LeaveCreateRequested>(_onCreateRequested);
    on<LeaveLoadMyList>(_onLoadMyList);
    on<LeaveLoadAdminList>(_onLoadAdminList);
    on<LeaveApproveRequested>(_onApproveRequested);
    on<LeaveRejectRequested>(_onRejectRequested);
    on<LeaveBatchApproveRequested>(_onBatchApproveRequested);
  }

  Future<void> _onCreateRequested(
    LeaveCreateRequested event,
    Emitter<LeaveState> emit,
  ) async {
    emit(LeaveLoading());
    try {
      final leave = await _leaveRepository.createLeave(
        leaveDate: event.leaveDate,
        leaveType: event.leaveType,
        description: event.description,
      );
      emit(LeaveCreateSuccess(leave));
    } catch (e) {
      emit(LeaveFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadMyList(
    LeaveLoadMyList event,
    Emitter<LeaveState> emit,
  ) async {
    emit(LeaveLoading());
    try {
      final leaves = await _leaveRepository.getMyLeaves(
        status: event.status,
      );
      emit(LeaveMyListLoaded(leaves));
    } catch (e) {
      emit(LeaveFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadAdminList(
    LeaveLoadAdminList event,
    Emitter<LeaveState> emit,
  ) async {
    emit(LeaveLoading());
    try {
      final leaves = await _leaveRepository.getAdminLeaves(
        status: event.status,
      );
      emit(LeaveAdminListLoaded(leaves));
    } catch (e) {
      emit(LeaveFailure(_extractMessage(e)));
    }
  }

  Future<void> _onApproveRequested(
    LeaveApproveRequested event,
    Emitter<LeaveState> emit,
  ) async {
    emit(LeaveLoading());
    try {
      final leave = await _leaveRepository.processLeave(
        leaveId: event.leaveId,
        status: 'approved',
        managerNote: event.managerNote,
      );
      emit(LeaveProcessSuccess(leave, 'Đã duyệt yêu cầu nghỉ phép'));
    } catch (e) {
      emit(LeaveFailure(_extractMessage(e)));
    }
  }

  Future<void> _onRejectRequested(
    LeaveRejectRequested event,
    Emitter<LeaveState> emit,
  ) async {
    emit(LeaveLoading());
    try {
      final leave = await _leaveRepository.processLeave(
        leaveId: event.leaveId,
        status: 'rejected',
        managerNote: event.managerNote,
      );
      emit(LeaveProcessSuccess(leave, 'Đã từ chối yêu cầu nghỉ phép'));
    } catch (e) {
      emit(LeaveFailure(_extractMessage(e)));
    }
  }

  Future<void> _onBatchApproveRequested(
    LeaveBatchApproveRequested event,
    Emitter<LeaveState> emit,
  ) async {
    emit(LeaveLoading());
    try {
      final count = await _leaveRepository.batchApprove();
      emit(LeaveBatchApproveSuccess(count));
    } catch (e) {
      emit(LeaveFailure(_extractMessage(e)));
    }
  }

  String _extractMessage(Object e) {
    if (e is DioException) {
      final data = e.response?.data;
      if (data is Map<String, dynamic>) {
        final error = data['error'];
        if (error is Map<String, dynamic> && error['message'] != null) {
          return error['message'].toString();
        }
      }
      return 'Lỗi kết nối. Vui lòng thử lại.';
    }
    return e.toString().replaceFirst('Exception: ', '');
  }
}
