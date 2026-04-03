import 'package:equatable/equatable.dart';
import '../../../data/models/attendance_model.dart';

abstract class AttendanceState extends Equatable {
  const AttendanceState();

  @override
  List<Object?> get props => [];
}

class AttendanceInitial extends AttendanceState {}

class AttendanceLoading extends AttendanceState {}

class AttendanceTodayLoaded extends AttendanceState {
  final AttendanceModel? today;

  const AttendanceTodayLoaded(this.today);

  @override
  List<Object?> get props => [today];
}

class AttendanceCheckInSuccess extends AttendanceState {
  final AttendanceModel attendance;
  final DateTime timestamp;

  AttendanceCheckInSuccess(this.attendance) : timestamp = DateTime.now();

  @override
  List<Object?> get props => [attendance, timestamp];
}

class AttendanceCheckOutSuccess extends AttendanceState {
  final AttendanceModel attendance;
  final DateTime timestamp;

  AttendanceCheckOutSuccess(this.attendance) : timestamp = DateTime.now();

  @override
  List<Object?> get props => [attendance, timestamp];
}

class AttendanceHistoryLoaded extends AttendanceState {
  final List<AttendanceModel> records;
  final bool hasMore;

  const AttendanceHistoryLoaded({required this.records, this.hasMore = true});

  @override
  List<Object?> get props => [records, hasMore];
}

class AttendanceFailure extends AttendanceState {
  final String message;

  const AttendanceFailure(this.message);

  @override
  List<Object?> get props => [message];
}
