import 'package:dio/dio.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../domain/repositories/holiday_repository.dart';
import 'holiday_event.dart';
import 'holiday_state.dart';

class HolidayBloc extends Bloc<HolidayEvent, HolidayState> {
  final HolidayRepository _holidayRepository;

  HolidayBloc({required HolidayRepository holidayRepository})
      : _holidayRepository = holidayRepository,
        super(HolidayInitial()) {
    on<HolidayLoadCalendar>(_onLoadCalendar);
    on<HolidayLoadSummary>(_onLoadSummary);
  }

  Future<void> _onLoadCalendar(
    HolidayLoadCalendar event,
    Emitter<HolidayState> emit,
  ) async {
    emit(HolidayLoading());
    try {
      final list = await _holidayRepository.getCalendar(
        dateFrom: event.dateFrom,
        dateTo: event.dateTo,
      );
      emit(HolidayCalendarLoaded(list));
    } catch (e) {
      emit(HolidayFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadSummary(
    HolidayLoadSummary event,
    Emitter<HolidayState> emit,
  ) async {
    emit(HolidayLoading());
    try {
      final summary = await _holidayRepository.getAttendanceSummary(
        dateFrom: event.dateFrom,
        dateTo: event.dateTo,
      );
      emit(HolidaySummaryLoaded(summary));
    } catch (e) {
      emit(HolidayFailure(_extractMessage(e)));
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
