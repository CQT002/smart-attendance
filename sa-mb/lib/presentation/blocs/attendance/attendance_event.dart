import 'package:equatable/equatable.dart';

abstract class AttendanceEvent extends Equatable {
  const AttendanceEvent();

  @override
  List<Object?> get props => [];
}

class AttendanceLoadToday extends AttendanceEvent {}

class AttendanceCheckInRequested extends AttendanceEvent {
  final String method; // 'wifi' or 'gps'

  const AttendanceCheckInRequested({required this.method});

  @override
  List<Object?> get props => [method];
}

class AttendanceCheckOutRequested extends AttendanceEvent {
  final int attendanceId;
  final String method;

  const AttendanceCheckOutRequested({
    required this.attendanceId,
    required this.method,
  });

  @override
  List<Object?> get props => [attendanceId, method];
}

class AttendanceLoadHistory extends AttendanceEvent {
  final DateTime from;
  final DateTime to;
  final int page;

  const AttendanceLoadHistory({
    required this.from,
    required this.to,
    this.page = 1,
  });

  @override
  List<Object?> get props => [from, to, page];
}
