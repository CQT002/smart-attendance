import 'package:equatable/equatable.dart';

abstract class HolidayEvent extends Equatable {
  const HolidayEvent();

  @override
  List<Object?> get props => [];
}

/// Load ngày lễ (mặc định năm hiện tại nếu bỏ trống)
class HolidayLoadCalendar extends HolidayEvent {
  final String? dateFrom;
  final String? dateTo;

  const HolidayLoadCalendar({this.dateFrom, this.dateTo});

  @override
  List<Object?> get props => [dateFrom, dateTo];
}

/// Load báo cáo công (kèm ngày lễ & hệ số)
class HolidayLoadSummary extends HolidayEvent {
  final String dateFrom;
  final String dateTo;

  const HolidayLoadSummary({required this.dateFrom, required this.dateTo});

  @override
  List<Object?> get props => [dateFrom, dateTo];
}
