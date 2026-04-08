import 'package:dio/dio.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../domain/repositories/overtime_repository.dart';
import 'overtime_event.dart';
import 'overtime_state.dart';

class OvertimeBloc extends Bloc<OvertimeEvent, OvertimeState> {
  final OvertimeRepository _overtimeRepository;

  OvertimeBloc({required OvertimeRepository overtimeRepository})
      : _overtimeRepository = overtimeRepository,
        super(OvertimeInitial()) {
    on<OvertimeCheckInRequested>(_onCheckInRequested);
    on<OvertimeCheckOutRequested>(_onCheckOutRequested);
    on<OvertimeLoadToday>(_onLoadToday);
    on<OvertimeLoadMyList>(_onLoadMyList);
    on<OvertimeLoadAdminList>(_onLoadAdminList);
    on<OvertimeApproveRequested>(_onApproveRequested);
    on<OvertimeRejectRequested>(_onRejectRequested);
    on<OvertimeBatchApproveRequested>(_onBatchApproveRequested);
  }

  Future<void> _onCheckInRequested(
    OvertimeCheckInRequested event,
    Emitter<OvertimeState> emit,
  ) async {
    emit(OvertimeLoading());
    try {
      final response = await _overtimeRepository.checkIn();
      emit(OvertimeCheckInSuccess(response));
    } catch (e) {
      emit(OvertimeFailure(_extractMessage(e)));
    }
  }

  Future<void> _onCheckOutRequested(
    OvertimeCheckOutRequested event,
    Emitter<OvertimeState> emit,
  ) async {
    emit(OvertimeLoading());
    try {
      final response = await _overtimeRepository.checkOut(
        overtimeId: event.overtimeId,
      );
      emit(OvertimeCheckOutSuccess(response));
    } catch (e) {
      emit(OvertimeFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadToday(
    OvertimeLoadToday event,
    Emitter<OvertimeState> emit,
  ) async {
    emit(OvertimeLoading());
    try {
      final overtime = await _overtimeRepository.getMyToday();
      emit(OvertimeTodayLoaded(overtime));
    } catch (e) {
      emit(OvertimeFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadMyList(
    OvertimeLoadMyList event,
    Emitter<OvertimeState> emit,
  ) async {
    emit(OvertimeLoading());
    try {
      final overtimes = await _overtimeRepository.getMyList(
        status: event.status,
      );
      emit(OvertimeMyListLoaded(overtimes));
    } catch (e) {
      emit(OvertimeFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadAdminList(
    OvertimeLoadAdminList event,
    Emitter<OvertimeState> emit,
  ) async {
    emit(OvertimeLoading());
    try {
      final overtimes = await _overtimeRepository.getAdminList(
        status: event.status,
      );
      emit(OvertimeAdminListLoaded(overtimes));
    } catch (e) {
      emit(OvertimeFailure(_extractMessage(e)));
    }
  }

  Future<void> _onApproveRequested(
    OvertimeApproveRequested event,
    Emitter<OvertimeState> emit,
  ) async {
    emit(OvertimeLoading());
    try {
      final overtime = await _overtimeRepository.process(
        overtimeId: event.overtimeId,
        status: 'approved',
        managerNote: event.managerNote,
      );
      emit(OvertimeProcessSuccess(overtime, 'Đã duyệt yêu cầu tăng ca'));
    } catch (e) {
      emit(OvertimeFailure(_extractMessage(e)));
    }
  }

  Future<void> _onRejectRequested(
    OvertimeRejectRequested event,
    Emitter<OvertimeState> emit,
  ) async {
    emit(OvertimeLoading());
    try {
      final overtime = await _overtimeRepository.process(
        overtimeId: event.overtimeId,
        status: 'rejected',
        managerNote: event.managerNote,
      );
      emit(OvertimeProcessSuccess(overtime, 'Đã từ chối yêu cầu tăng ca'));
    } catch (e) {
      emit(OvertimeFailure(_extractMessage(e)));
    }
  }

  Future<void> _onBatchApproveRequested(
    OvertimeBatchApproveRequested event,
    Emitter<OvertimeState> emit,
  ) async {
    emit(OvertimeLoading());
    try {
      final count = await _overtimeRepository.batchApprove();
      emit(OvertimeBatchApproveSuccess(count));
    } catch (e) {
      emit(OvertimeFailure(_extractMessage(e)));
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
