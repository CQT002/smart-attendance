import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/attendance_model.dart';
import '../blocs/attendance/attendance_bloc.dart';
import '../blocs/attendance/attendance_event.dart';
import '../blocs/attendance/attendance_state.dart';
import '../widgets/attendance_card.dart';

class HistoryScreen extends StatefulWidget {
  const HistoryScreen({super.key});

  @override
  State<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends State<HistoryScreen> {
  late DateTime _currentMonth;
  AttendanceModel? _selectedDayRecord;
  DateTime? _selectedDay;
  List<AttendanceModel> _cachedRecords = [];
  bool _isLoadingHistory = false;

  @override
  void initState() {
    super.initState();
    _currentMonth = DateTime(DateTime.now().year, DateTime.now().month, 1);
    _loadMonth();
  }

  void _loadMonth() {
    final from = AppDateUtils.startOfMonth(_currentMonth);
    final to = AppDateUtils.endOfMonth(_currentMonth);
    context.read<AttendanceBloc>().add(
          AttendanceLoadHistory(from: from, to: to),
        );
    setState(() {
      _selectedDayRecord = null;
      _selectedDay = null;
    });
  }

  void _previousMonth() {
    setState(() {
      _currentMonth = DateTime(_currentMonth.year, _currentMonth.month - 1, 1);
    });
    _loadMonth();
  }

  void _nextMonth() {
    setState(() {
      _currentMonth = DateTime(_currentMonth.year, _currentMonth.month + 1, 1);
    });
    _loadMonth();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SafeArea(
      child: Column(
        children: [
          // Header
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 16, 20, 0),
            child: Text(
              'Lịch sử chấm công',
              style: theme.textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Month navigation
          _buildMonthNav(theme),
          const SizedBox(height: 8),

          // Calendar + detail
          Expanded(
            child: BlocListener<AttendanceBloc, AttendanceState>(
              listener: (context, state) {
                if (state is AttendanceLoading && _isLoadingHistory) {
                  // Chỉ show loading khi đang load history, không phải check-in/out
                } else if (state is AttendanceHistoryLoaded) {
                  setState(() {
                    _isLoadingHistory = false;
                    _cachedRecords = state.records;
                  });
                } else if (state is AttendanceCheckInSuccess || state is AttendanceCheckOutSuccess) {
                  // Sau check-in/checkout thành công → reload history để cập nhật data mới nhất
                  _loadMonth();
                }
              },
              child: Builder(
                builder: (context) {
                  if (_isLoadingHistory && _cachedRecords.isEmpty) {
                    return const Center(child: CircularProgressIndicator());
                  }

                  final records = _cachedRecords;
                  final recordMap = <String, AttendanceModel>{};
                  for (final r in records) {
                    recordMap[_dateKey(r.date)] = r;
                  }

                  return SingleChildScrollView(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    child: Column(
                      children: [
                        _buildCalendar(theme, recordMap),
                        const SizedBox(height: 16),
                        _buildLegend(theme),
                        const SizedBox(height: 16),
                        _buildSummary(theme, records),
                        const SizedBox(height: 16),
                        if (_selectedDayRecord != null)
                          AttendanceCard(attendance: _selectedDayRecord!),
                        if (_selectedDay != null && _selectedDayRecord == null)
                          _buildNoRecordCard(theme),
                        const SizedBox(height: 24),
                      ],
                    ),
                  );
                },
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMonthNav(ThemeData theme) {
    final now = DateTime.now();
    final monthNames = [
      '', 'Tháng 1', 'Tháng 2', 'Tháng 3', 'Tháng 4', 'Tháng 5', 'Tháng 6',
      'Tháng 7', 'Tháng 8', 'Tháng 9', 'Tháng 10', 'Tháng 11', 'Tháng 12',
    ];
    final canGoNext = _currentMonth.isBefore(DateTime(now.year, now.month, 1));

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Card(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              IconButton(
                icon: const Icon(Icons.chevron_left_rounded),
                onPressed: _previousMonth,
                color: AppColors.primary,
              ),
              Text(
                '${monthNames[_currentMonth.month]} - ${_currentMonth.year}',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: AppColors.primary,
                ),
              ),
              IconButton(
                icon: const Icon(Icons.chevron_right_rounded),
                onPressed: canGoNext ? _nextMonth : null,
                color: AppColors.primary,
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCalendar(ThemeData theme, Map<String, AttendanceModel> recordMap) {
    final firstDay = DateTime(_currentMonth.year, _currentMonth.month, 1);
    final daysInMonth = DateTime(_currentMonth.year, _currentMonth.month + 1, 0).day;
    // Convert Monday=1..Sunday=7 to Sunday-first: Sunday=0, Monday=1..Saturday=6
    final startWeekday = firstDay.weekday % 7; // Sun=0, Mon=1, ..., Sat=6
    final today = DateTime.now();

    final dayLabels = ['CN', 'T2', 'T3', 'T4', 'T5', 'T6', 'T7'];

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          children: [
            // Day headers
            Row(
              children: dayLabels.asMap().entries.map((e) {
                final isWeekend = e.key == 0 || e.key == 6; // CN or T7
                return Expanded(
                  child: Center(
                    child: Text(
                      e.value,
                      style: theme.textTheme.bodySmall?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: isWeekend ? AppColors.error.withOpacity(0.7) : AppColors.primary,
                      ),
                    ),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 8),

            // Day grid
            ..._buildWeeks(firstDay, daysInMonth, startWeekday, today, recordMap, theme),
          ],
        ),
      ),
    );
  }

  List<Widget> _buildWeeks(
    DateTime firstDay,
    int daysInMonth,
    int startWeekday,
    DateTime today,
    Map<String, AttendanceModel> recordMap,
    ThemeData theme,
  ) {
    final weeks = <Widget>[];
    var dayCounter = 1;
    // startWeekday: Sun=0, Mon=1, ..., Sat=6
    final totalCells = startWeekday + daysInMonth;
    final totalWeeks = (totalCells / 7).ceil();

    for (var week = 0; week < totalWeeks; week++) {
      final cells = <Widget>[];
      for (var col = 0; col < 7; col++) {
        final cellIndex = week * 7 + col;
        final dayOffset = cellIndex - startWeekday;

        if (dayOffset < 0 || dayOffset >= daysInMonth) {
          cells.add(const Expanded(child: SizedBox(height: 44)));
          continue;
        }

        final day = dayCounter;
        dayCounter++;
        final date = DateTime(_currentMonth.year, _currentMonth.month, day);
        final key = _dateKey(date);
        final record = recordMap[key];
        final isToday = date.year == today.year && date.month == today.month && date.day == today.day;
        final isFuture = date.isAfter(today);
        final isWeekend = col == 0 || col == 6; // CN (col 0) hoặc T7 (col 6)
        final isSelected = _selectedDay != null &&
            _selectedDay!.year == date.year &&
            _selectedDay!.month == date.month &&
            _selectedDay!.day == date.day;

        final isHoliday = _isHoliday(date);
        final color = _getDayColor(record, isFuture, isWeekend, isHoliday);

        cells.add(
          Expanded(
            child: GestureDetector(
              onTap: (isFuture && !isToday)
                  ? null
                  : () {
                      setState(() {
                        _selectedDay = date;
                        _selectedDayRecord = record;
                      });
                    },
              child: Container(
                height: 40,
                width: 40,
                margin: const EdgeInsets.all(2),
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isSelected ? color : null,
                  border: color != null
                      ? Border.all(
                          color: isToday ? AppColors.primary : color,
                          width: isToday ? 2.5 : 2,
                        )
                      : isToday
                          ? Border.all(color: AppColors.primary, width: 2.5)
                          : null,
                ),
                child: Center(
                  child: Text(
                    '$day',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      fontWeight: isToday ? FontWeight.w800 : FontWeight.w600,
                      color: isSelected && color != null
                          ? Colors.white
                          : color ?? AppColors.textPrimary,
                    ),
                  ),
                ),
              ),
            ),
          ),
        );
      }
      weeks.add(
        Padding(
          padding: const EdgeInsets.only(bottom: 2),
          child: Row(children: cells),
        ),
      );
    }
    return weeks;
  }

  Color? _getDayColor(AttendanceModel? record, bool isFuture, bool isWeekend, bool isHoliday) {
    // Ngày tương lai — trắng (null)
    if (isFuture) return null;

    // Cuối tuần (T7/CN) — xám
    if (isWeekend && record == null) return AppColors.calendarDayOff;

    // Ngày lễ trong tuần (T2-T6) — xanh lá (được nghỉ, vẫn tính công)
    if (isHoliday && !isWeekend) return AppColors.calendarPresent;

    if (record == null) {
      return AppColors.calendarAbsent; // Đỏ — chưa chấm công ngày thường
    }

    // Nghỉ có phép
    if (record.status == 'leave' || record.status == 'approved_leave') {
      return AppColors.calendarLeave;
    }

    // Đã check-in và check-out đầy đủ
    if (record.hasCheckedIn && record.hasCheckedOut) {
      if (record.workHours >= 8) {
        return AppColors.calendarPresent; // Xanh lá — đủ công
      }
      return AppColors.calendarIncomplete; // Cam — < 8h
    }

    // Chỉ check-in hoặc chỉ check-out
    if (record.hasCheckedIn || record.hasCheckedOut) {
      return AppColors.calendarIncomplete; // Cam — thiếu
    }

    return AppColors.calendarAbsent; // Đỏ
  }

  /// Ngày lễ Việt Nam (dương lịch cố định).
  /// Âm lịch (Tết, Giỗ Tổ...) cần tính riêng hoặc lấy từ API.
  bool _isHoliday(DateTime date) {
    final md = (date.month, date.day);
    return const {
      (1, 1),   // Tết Dương lịch
      (4, 30),  // Giải phóng miền Nam
      (5, 1),   // Quốc tế Lao động
      (9, 2),   // Quốc khánh
    }.contains(md);
  }

  Widget _buildLegend(ThemeData theme) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Wrap(
          spacing: 16,
          runSpacing: 8,
          alignment: WrapAlignment.center,
          children: [
            _legendItem(theme, AppColors.calendarPresent, 'Đủ công'),
            _legendItem(theme, AppColors.calendarIncomplete, 'Thiếu'),
            _legendItem(theme, AppColors.calendarAbsent, 'Vắng'),
            _legendItem(theme, AppColors.calendarLeave, 'Nghỉ phép'),
            _legendItem(theme, AppColors.calendarDayOff, 'Cuối tuần'),
          ],
        ),
      ),
    );
  }

  Widget _legendItem(ThemeData theme, Color color, String label) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 14,
          height: 14,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            border: Border.all(color: color, width: 2),
          ),
        ),
        const SizedBox(width: 6),
        Text(
          label,
          style: theme.textTheme.bodySmall?.copyWith(
            color: AppColors.textSecondary,
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }

  Widget _buildSummary(ThemeData theme, List<AttendanceModel> records) {
    final now = DateTime.now();
    final daysInMonth = DateTime(_currentMonth.year, _currentMonth.month + 1, 0).day;
    final isCurrentMonth = _currentMonth.year == now.year && _currentMonth.month == now.month;
    final maxDay = isCurrentMonth ? now.day : daysInMonth;

    var workDays = 0;
    for (var d = 1; d <= maxDay; d++) {
      final date = DateTime(_currentMonth.year, _currentMonth.month, d);
      if (date.weekday < 6) workDays++;
    }

    // Đếm ngày lễ trong tuần (T2-T6) → tính đủ công
    var holidayWorkDays = 0;
    for (var d = 1; d <= maxDay; d++) {
      final date = DateTime(_currentMonth.year, _currentMonth.month, d);
      if (date.weekday < 6 && _isHoliday(date)) holidayWorkDays++;
    }

    final fullDays = records.where((r) => r.hasCheckedIn && r.hasCheckedOut && r.workHours >= 8).length + holidayWorkDays;
    final incompleteDays = records.where((r) => (r.hasCheckedIn || r.hasCheckedOut) && !(r.hasCheckedIn && r.hasCheckedOut && r.workHours >= 8)).length;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Text(
              'Tổng kết tháng',
              style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                _summaryItem(theme, '$fullDays', 'Đủ công', AppColors.calendarPresent),
                _summaryItem(theme, '$incompleteDays', 'Thiếu', AppColors.calendarIncomplete),
                _summaryItem(theme, '${workDays - fullDays - incompleteDays}', 'Vắng', AppColors.calendarAbsent),
                _summaryItem(theme, '$workDays', 'Ngày làm', AppColors.primary),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _summaryItem(ThemeData theme, String value, String label, Color color) {
    return Expanded(
      child: Column(
        children: [
          Text(
            value,
            style: theme.textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.w800,
              color: color,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            label,
            style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
          ),
        ],
      ),
    );
  }

  Widget _buildNoRecordCard(ThemeData theme) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            const Icon(Icons.event_busy, color: AppColors.textSecondary),
            const SizedBox(width: 12),
            Text(
              'Không có dữ liệu chấm công ngày ${AppDateUtils.formatDate(_selectedDay!)}',
              style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
            ),
          ],
        ),
      ),
    );
  }

  String _dateKey(DateTime date) => '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
}
