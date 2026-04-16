import 'package:equatable/equatable.dart';
import '../../../data/models/holiday_model.dart';

abstract class HolidayState extends Equatable {
  const HolidayState();

  @override
  List<Object?> get props => [];
}

class HolidayInitial extends HolidayState {}

class HolidayLoading extends HolidayState {}

class HolidayCalendarLoaded extends HolidayState {
  final List<HolidayModel> holidays;

  const HolidayCalendarLoaded(this.holidays);

  @override
  List<Object?> get props => [holidays];
}

class HolidaySummaryLoaded extends HolidayState {
  final AttendanceSummaryModel summary;

  const HolidaySummaryLoaded(this.summary);

  @override
  List<Object?> get props => [summary];
}

class HolidayFailure extends HolidayState {
  final String message;

  const HolidayFailure(this.message);

  @override
  List<Object?> get props => [message];
}
