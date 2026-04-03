import 'package:intl/intl.dart';

class AppDateUtils {
  static final DateFormat _dateFormat = DateFormat('dd/MM/yyyy');
  static final DateFormat _timeFormat = DateFormat('HH:mm');
  static final DateFormat _dateTimeFormat = DateFormat('dd/MM/yyyy HH:mm');
  static final DateFormat _monthYearFormat = DateFormat('MM/yyyy');
  static final DateFormat _dayNameFormat = DateFormat('EEEE', 'vi');

  /// Parse ISO8601 string từ API (có thể UTC) → DateTime local (HCM)
  static DateTime parseFromApi(String dateStr) {
    return DateTime.parse(dateStr).toLocal();
  }

  /// Parse ISO8601 string nullable
  static DateTime? parseFromApiNullable(String? dateStr) {
    if (dateStr == null || dateStr.isEmpty) return null;
    return DateTime.parse(dateStr).toLocal();
  }

  static String formatDate(DateTime date) => _dateFormat.format(date.toLocal());
  static String formatTime(DateTime date) => _timeFormat.format(date.toLocal());
  static String formatDateTime(DateTime date) => _dateTimeFormat.format(date.toLocal());
  static String formatMonthYear(DateTime date) => _monthYearFormat.format(date.toLocal());
  static String formatDayName(DateTime date) => _dayNameFormat.format(date.toLocal());

  static DateTime startOfDay(DateTime date) {
    return DateTime(date.year, date.month, date.day);
  }

  static DateTime endOfDay(DateTime date) {
    return DateTime(date.year, date.month, date.day, 23, 59, 59);
  }

  static DateTime startOfWeek(DateTime date) {
    final diff = date.weekday - DateTime.monday;
    return startOfDay(date.subtract(Duration(days: diff)));
  }

  static DateTime endOfWeek(DateTime date) {
    final diff = DateTime.sunday - date.weekday;
    return endOfDay(date.add(Duration(days: diff)));
  }

  static DateTime startOfMonth(DateTime date) {
    return DateTime(date.year, date.month, 1);
  }

  static DateTime endOfMonth(DateTime date) {
    return DateTime(date.year, date.month + 1, 0, 23, 59, 59);
  }

  static String formatWorkHours(double hours) {
    final h = hours.floor();
    final m = ((hours - h) * 60).round();
    if (m == 0) return '${h}h';
    return '${h}h${m}m';
  }
}
