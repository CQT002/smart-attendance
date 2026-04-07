import 'package:dio/dio.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../domain/repositories/correction_repository.dart';
import 'correction_event.dart';
import 'correction_state.dart';

class CorrectionBloc extends Bloc<CorrectionEvent, CorrectionState> {
  final CorrectionRepository _correctionRepository;

  CorrectionBloc({required CorrectionRepository correctionRepository})
      : _correctionRepository = correctionRepository,
        super(CorrectionInitial()) {
    on<CorrectionCreateRequested>(_onCreateRequested);
    on<CorrectionLoadMyList>(_onLoadMyList);
    on<CorrectionLoadAdminList>(_onLoadAdminList);
    on<CorrectionApproveRequested>(_onApproveRequested);
    on<CorrectionRejectRequested>(_onRejectRequested);
  }

  Future<void> _onCreateRequested(
    CorrectionCreateRequested event,
    Emitter<CorrectionState> emit,
  ) async {
    emit(CorrectionLoading());
    try {
      final correction = await _correctionRepository.createCorrection(
        attendanceLogId: event.attendanceLogId,
        description: event.description,
      );
      emit(CorrectionCreateSuccess(correction));
    } catch (e) {
      emit(CorrectionFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadMyList(
    CorrectionLoadMyList event,
    Emitter<CorrectionState> emit,
  ) async {
    emit(CorrectionLoading());
    try {
      final corrections = await _correctionRepository.getMyCorrections(
        status: event.status,
      );
      emit(CorrectionMyListLoaded(corrections));
    } catch (e) {
      emit(CorrectionFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadAdminList(
    CorrectionLoadAdminList event,
    Emitter<CorrectionState> emit,
  ) async {
    emit(CorrectionLoading());
    try {
      final corrections = await _correctionRepository.getAdminCorrections(
        status: event.status,
      );
      emit(CorrectionAdminListLoaded(corrections));
    } catch (e) {
      emit(CorrectionFailure(_extractMessage(e)));
    }
  }

  Future<void> _onApproveRequested(
    CorrectionApproveRequested event,
    Emitter<CorrectionState> emit,
  ) async {
    emit(CorrectionLoading());
    try {
      final correction = await _correctionRepository.processCorrection(
        correctionId: event.correctionId,
        status: 'approved',
        managerNote: event.managerNote,
      );
      emit(CorrectionProcessSuccess(correction, 'Đã duyệt yêu cầu bù công'));
    } catch (e) {
      emit(CorrectionFailure(_extractMessage(e)));
    }
  }

  Future<void> _onRejectRequested(
    CorrectionRejectRequested event,
    Emitter<CorrectionState> emit,
  ) async {
    emit(CorrectionLoading());
    try {
      final correction = await _correctionRepository.processCorrection(
        correctionId: event.correctionId,
        status: 'rejected',
        managerNote: event.managerNote,
      );
      emit(CorrectionProcessSuccess(correction, 'Đã từ chối yêu cầu bù công'));
    } catch (e) {
      emit(CorrectionFailure(_extractMessage(e)));
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
