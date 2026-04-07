import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/constants/app_constants.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/attendance_model.dart';
import '../../data/models/correction_model.dart';
import '../../data/models/leave_model.dart';
import '../../domain/repositories/correction_repository.dart';
import '../../domain/repositories/leave_repository.dart';
import '../blocs/attendance/attendance_bloc.dart';
import '../blocs/attendance/attendance_event.dart';
import '../blocs/attendance/attendance_state.dart';
import '../blocs/correction/correction_bloc.dart';
import '../blocs/leave/leave_bloc.dart';
import 'correction_request_screen.dart';
import 'leave_request_screen.dart';

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
  Map<int, CorrectionModel> _correctionsByLogId = {}; // attendance_log_id → correction
  Map<String, LeaveModel> _leavesByDate = {}; // "yyyy-MM-dd" → leave
  bool _isLoadingHistory = false;
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _currentMonth = DateTime(DateTime.now().year, DateTime.now().month, 1);
    _loadMonth();
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _loadMonth() {
    final from = AppDateUtils.startOfMonth(_currentMonth);
    final to = AppDateUtils.endOfMonth(_currentMonth);
    context.read<AttendanceBloc>().add(
          AttendanceLoadHistory(from: from, to: to),
        );
    _loadCorrections();
    _loadLeaves();
    setState(() {
      _selectedDayRecord = null;
      _selectedDay = null;
    });
  }

  Future<void> _loadCorrections() async {
    try {
      final repo = context.read<CorrectionRepository>();
      final corrections = await repo.getMyCorrections(limit: 100);
      setState(() {
        _correctionsByLogId = {
          for (final c in corrections) c.attendanceLogId: c,
        };
      });
    } catch (_) {
      // Fail silently — correction info is optional
    }
  }

  Future<void> _loadLeaves() async {
    try {
      final repo = context.read<LeaveRepository>();
      final leaves = await repo.getMyLeaves(limit: 100);
      setState(() {
        _leavesByDate = {
          for (final l in leaves)
            l.leaveDate.length >= 10 ? l.leaveDate.substring(0, 10) : l.leaveDate: l,
        };
      });
    } catch (_) {
      // Fail silently — leave info is optional
    }
  }

  void _selectDay(DateTime date, AttendanceModel? record) {
    setState(() {
      _selectedDay = date;
      _selectedDayRecord = record;
    });
    if (record != null) {
      _showDayDetail(record);
    } else {
      // Ngày không có record — cho phép đăng ký nghỉ phép (ngày tương lai hoặc ngày vắng chưa có log)
      _showLeaveOnlySheet(date);
    }
  }

  void _showDayDetail(AttendanceModel record) {
    final theme = Theme.of(context);
    final correctable = _isCorrectable(record);
    final existingCorrection = _correctionsByLogId[record.id];
    final existingLeave = _leavesByDate[_dateKey(record.date)];

    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Handle bar
            Container(
              width: 40, height: 4,
              margin: const EdgeInsets.only(bottom: 16),
              decoration: BoxDecoration(
                color: Colors.grey[300],
                borderRadius: BorderRadius.circular(2),
              ),
            ),

            // Dòng 1: Ngày + Badge
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  AppDateUtils.formatDate(record.date),
                  style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700),
                ),
                _buildStatusBadge(record),
              ],
            ),
            const SizedBox(height: 16),

            // Dòng 2: Check-in / Check-out / Giờ làm
            Row(
              children: [
                _buildTimeInfo(theme, icon: Icons.login_rounded, label: 'Vào',
                  time: record.checkInTime != null ? AppDateUtils.formatTime(record.checkInTime!) : '--:--',
                  method: record.checkInMethod, color: AppColors.success),
                const SizedBox(width: 24),
                _buildTimeInfo(theme, icon: Icons.logout_rounded, label: 'Ra',
                  time: record.checkOutTime != null ? AppDateUtils.formatTime(record.checkOutTime!) : '--:--',
                  method: record.checkOutMethod, color: AppColors.error),
                const Spacer(),
                if (record.workHours > 0)
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(AppDateUtils.formatWorkHours(record.workHours),
                        style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold, color: AppColors.primary)),
                      Text('Giờ làm', style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
                    ],
                  ),
              ],
            ),

            // Dòng 3: Trạng thái correction hoặc Button tạo mới
            if (existingCorrection != null) ...[
              const SizedBox(height: 16),
              _buildCorrectionStatus(theme, existingCorrection),
            ] else if (correctable) ...[
              const SizedBox(height: 20),
              Row(
                children: [
                  Expanded(
                    child: SizedBox(
                      height: 48,
                      child: ElevatedButton.icon(
                        onPressed: () {
                          Navigator.pop(ctx);
                          _openCorrectionRequest();
                        },
                        icon: const Icon(Icons.edit_calendar_outlined, size: 18),
                        label: const Text('Bù công',
                          style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.primary,
                          foregroundColor: Colors.white,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ],
            // Trạng thái nghỉ phép (nếu đã có đơn)
            if (existingLeave != null) ...[
              const SizedBox(height: 16),
              _buildLeaveStatus(theme, existingLeave),
            ]
            // Nút nghỉ phép cho ngày quá khứ leavable (absent, half_day) — chỉ hiện khi chưa có đơn
            else if (existingCorrection == null && _isLeavable(record)) ...[
              const SizedBox(height: 8),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: OutlinedButton.icon(
                  onPressed: () {
                    Navigator.pop(ctx);
                    _openLeaveRequest(record.date, record);
                  },
                  icon: const Icon(Icons.event_busy_outlined, size: 18),
                  label: const Text('Đăng ký nghỉ phép',
                    style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: AppColors.calendarLeave,
                    side: const BorderSide(color: AppColors.calendarLeave),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildCorrectionStatus(ThemeData theme, CorrectionModel correction) {
    Color statusColor;
    IconData statusIcon;
    String statusText;

    switch (correction.status) {
      case 'approved':
        statusColor = AppColors.success;
        statusIcon = Icons.check_circle_outline;
        statusText = 'Đã duyệt bù công';
        break;
      case 'rejected':
        statusColor = AppColors.error;
        statusIcon = Icons.cancel_outlined;
        statusText = 'Từ chối bù công';
        break;
      default: // pending
        statusColor = AppColors.warning;
        statusIcon = Icons.hourglass_top_rounded;
        statusText = 'Đang đợi duyệt';
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: statusColor.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: statusColor.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(statusIcon, color: statusColor, size: 20),
              const SizedBox(width: 8),
              Text(statusText,
                style: theme.textTheme.titleSmall?.copyWith(
                  color: statusColor, fontWeight: FontWeight.w700)),
            ],
          ),
          const SizedBox(height: 8),
          Text('Lý do: ${correction.description}',
            style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
          if (correction.managerNote.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text('Ghi chú: ${correction.managerNote}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: AppColors.textSecondary, fontStyle: FontStyle.italic)),
          ],
          if (correction.processedBy != null) ...[
            const SizedBox(height: 4),
            Text('Người duyệt: ${correction.processedBy!.name}',
              style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
          ],
          if (correction.processedAt != null) ...[
            const SizedBox(height: 4),
            Text('Thời gian: ${AppDateUtils.formatDateTime(correction.processedAt!)}',
              style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
          ],
        ],
      ),
    );
  }

  Widget _buildLeaveStatus(ThemeData theme, LeaveModel leave) {
    Color statusColor;
    IconData statusIcon;
    String statusText;

    switch (leave.status) {
      case 'approved':
        statusColor = AppColors.success;
        statusIcon = Icons.check_circle_outline;
        statusText = 'Đã duyệt nghỉ phép';
        break;
      case 'rejected':
        statusColor = AppColors.error;
        statusIcon = Icons.cancel_outlined;
        statusText = 'Từ chối nghỉ phép';
        break;
      default: // pending
        statusColor = AppColors.warning;
        statusIcon = Icons.hourglass_top_rounded;
        statusText = 'Nghỉ phép đang chờ duyệt';
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: statusColor.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: statusColor.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(statusIcon, color: statusColor, size: 20),
              const SizedBox(width: 8),
              Text(statusText,
                style: theme.textTheme.titleSmall?.copyWith(
                  color: statusColor, fontWeight: FontWeight.w700)),
            ],
          ),
          const SizedBox(height: 8),
          Text('${leave.leaveTypeDisplay} (${leave.timeRangeDisplay})',
            style: theme.textTheme.bodySmall?.copyWith(fontWeight: FontWeight.w600)),
          const SizedBox(height: 4),
          Text('Lý do: ${leave.description}',
            style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
          if (leave.managerNote.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text('Ghi chú: ${leave.managerNote}',
              style: theme.textTheme.bodySmall?.copyWith(
                color: AppColors.textSecondary, fontStyle: FontStyle.italic)),
          ],
          if (leave.processedBy != null) ...[
            const SizedBox(height: 4),
            Text('Người duyệt: ${leave.processedBy!.name}',
              style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
          ],
          if (leave.processedAt != null) ...[
            const SizedBox(height: 4),
            Text('Thời gian: ${AppDateUtils.formatDateTime(leave.processedAt!)}',
              style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
          ],
        ],
      ),
    );
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
                if (state is AttendanceHistoryLoaded) {
                  setState(() {
                    _isLoadingHistory = false;
                    _cachedRecords = state.records;
                  });
                } else if (state is AttendanceCheckInSuccess || state is AttendanceCheckOutSuccess) {
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
                    controller: _scrollController,
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    child: Column(
                      children: [
                        _buildCalendar(theme, recordMap),
                        const SizedBox(height: 16),
                        _buildLegend(theme),
                        const SizedBox(height: 16),
                        _buildSummary(theme, records),
                        if (_selectedDay != null && _selectedDayRecord == null) ...[
                          const SizedBox(height: 16),
                          _buildNoRecordCard(theme),
                        ],
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

  Widget _buildStatusBadge(AttendanceModel record) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: _statusColor(record.status).withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _statusColor(record.status).withValues(alpha: 0.4)),
      ),
      child: Text(
        record.statusDisplay,
        style: TextStyle(
          color: _statusColor(record.status),
          fontSize: 12,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'present':
        return AppColors.success;
      case 'late':
      case 'early_leave':
      case 'late_early_leave':
      case 'half_day':
        return AppColors.warning;
      case 'absent':
        return AppColors.error;
      case 'leave':
      case 'half_day_leave':
        return AppColors.calendarLeave;
      default:
        return AppColors.textSecondary;
    }
  }

  Widget _buildTimeInfo(
    ThemeData theme, {
    required IconData icon,
    required String label,
    required String time,
    String? method,
    required Color color,
  }) {
    return Row(
      children: [
        Icon(icon, size: 20, color: color),
        const SizedBox(width: 8),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label, style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
            Text(time, style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600)),
            if (method != null)
              Row(
                children: [
                  Icon(method == 'wifi' ? Icons.wifi : Icons.gps_fixed, size: 12, color: AppColors.textSecondary),
                  const SizedBox(width: 2),
                  Text(
                    method == 'wifi' ? 'WiFi' : 'GPS',
                    style: theme.textTheme.bodySmall?.copyWith(fontSize: 10, color: AppColors.textSecondary),
                  ),
                ],
              ),
          ],
        ),
      ],
    );
  }

  Widget _buildMonthNav(ThemeData theme) {
    final now = DateTime.now();
    final monthNames = [
      '', 'Tháng 1', 'Tháng 2', 'Tháng 3', 'Tháng 4', 'Tháng 5', 'Tháng 6',
      'Tháng 7', 'Tháng 8', 'Tháng 9', 'Tháng 10', 'Tháng 11', 'Tháng 12',
    ];
    // Cho phép xem tháng tới để đăng ký nghỉ phép tương lai
    final canGoNext = _currentMonth.isBefore(DateTime(now.year, now.month + 1, 1));

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
    final startWeekday = firstDay.weekday % 7;
    final today = DateTime.now();

    final dayLabels = ['CN', 'T2', 'T3', 'T4', 'T5', 'T6', 'T7'];

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          children: [
            Row(
              children: dayLabels.asMap().entries.map((e) {
                final isWeekend = e.key == 0 || e.key == 6;
                return Expanded(
                  child: Center(
                    child: Text(
                      e.value,
                      style: theme.textTheme.bodySmall?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: isWeekend ? AppColors.error.withValues(alpha: 0.7) : AppColors.primary,
                      ),
                    ),
                  ),
                );
              }).toList(),
            ),
            const SizedBox(height: 8),
            ..._buildWeeks(firstDay, daysInMonth, startWeekday, today, recordMap, theme),
          ],
        ),
      ),
    );
  }

  List<Widget> _buildWeeks(
    DateTime firstDay, int daysInMonth, int startWeekday,
    DateTime today, Map<String, AttendanceModel> recordMap, ThemeData theme,
  ) {
    final weeks = <Widget>[];
    var dayCounter = 1;
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
        final isWeekend = col == 0 || col == 6;
        final isSelected = _selectedDay != null &&
            _selectedDay!.year == date.year &&
            _selectedDay!.month == date.month &&
            _selectedDay!.day == date.day;

        final isHoliday = _isHoliday(date);
        final color = _getDayColor(record, isFuture, isWeekend, isHoliday);

        cells.add(
          Expanded(
            child: GestureDetector(
              onTap: () => _selectDay(date, record),
              child: Container(
                height: 40,
                width: 40,
                margin: const EdgeInsets.all(2),
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isSelected ? color : null,
                  border: color != null
                      ? Border.all(color: isToday ? AppColors.primary : color, width: isToday ? 2.5 : 2)
                      : isToday
                          ? Border.all(color: AppColors.primary, width: 2.5)
                          : null,
                ),
                child: Center(
                  child: Text(
                    '$day',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      fontWeight: isToday ? FontWeight.w800 : FontWeight.w600,
                      color: isSelected && color != null ? Colors.white : color ?? AppColors.textPrimary,
                    ),
                  ),
                ),
              ),
            ),
          ),
        );
      }
      weeks.add(Padding(padding: const EdgeInsets.only(bottom: 2), child: Row(children: cells)));
    }
    return weeks;
  }

  Color? _getDayColor(AttendanceModel? record, bool isFuture, bool isWeekend, bool isHoliday) {
    if (isFuture) return null;
    if (isWeekend && record == null) return AppColors.calendarDayOff;
    if (isHoliday && !isWeekend) return AppColors.calendarPresent;
    if (record == null) return AppColors.calendarAbsent;
    if (record.status == 'leave' || record.status == 'approved_leave') return AppColors.calendarLeave;
    if (record.hasCheckedIn && record.hasCheckedOut) {
      return record.workHours >= 8 ? AppColors.calendarPresent : AppColors.calendarIncomplete;
    }
    if (record.hasCheckedIn || record.hasCheckedOut) return AppColors.calendarIncomplete;
    return AppColors.calendarAbsent;
  }

  bool _isHoliday(DateTime date) {
    final md = (date.month, date.day);
    return const {(1, 1), (4, 30), (5, 1), (9, 2)}.contains(md);
  }

  Widget _buildLegend(ThemeData theme) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Wrap(
          spacing: 16, runSpacing: 8, alignment: WrapAlignment.center,
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
        Container(width: 14, height: 14, decoration: BoxDecoration(shape: BoxShape.circle, border: Border.all(color: color, width: 2))),
        const SizedBox(width: 6),
        Text(label, style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary, fontWeight: FontWeight.w500)),
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
    var holidayWorkDays = 0;
    for (var d = 1; d <= maxDay; d++) {
      final date = DateTime(_currentMonth.year, _currentMonth.month, d);
      if (date.weekday < 6 && _isHoliday(date)) holidayWorkDays++;
    }

    // Nghỉ phép đã duyệt (leave, half_day_leave) tính là đủ công
    const leaveStatuses = {'leave', 'half_day_leave'};

    final fullDays = records.where((r) =>
        leaveStatuses.contains(r.status) ||
        (r.hasCheckedIn && r.hasCheckedOut && r.workHours >= 8)).length + holidayWorkDays;

    final incompleteDays = records.where((r) =>
        !leaveStatuses.contains(r.status) &&
        (r.hasCheckedIn || r.hasCheckedOut) &&
        !(r.hasCheckedIn && r.hasCheckedOut && r.workHours >= 8)).length;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Text('Tổng kết tháng', style: theme.textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700)),
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
          Text(value, style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800, color: color)),
          const SizedBox(height: 4),
          Text(label, style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
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

  bool _isLeavable(AttendanceModel record) {
    const leavableStatuses = {
      AppConstants.statusAbsent,
      AppConstants.statusHalfDay,
    };
    return leavableStatuses.contains(record.status);
  }

  void _openLeaveRequest(DateTime date, AttendanceModel? attendance) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => BlocProvider.value(
          value: context.read<LeaveBloc>(),
          child: LeaveRequestScreen(
            selectedDate: date,
            attendance: attendance,
          ),
        ),
      ),
    );
    if (result == true) {
      _loadMonth();
    }
  }

  void _showLeaveOnlySheet(DateTime date) {
    final theme = Theme.of(context);
    final today = DateTime.now();
    final isWeekend = date.weekday == 6 || date.weekday == 7;

    // Không hiện sheet cho cuối tuần
    if (isWeekend) return;

    final existingLeave = _leavesByDate[_dateKey(date)];

    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 40, height: 4,
              margin: const EdgeInsets.only(bottom: 16),
              decoration: BoxDecoration(
                color: Colors.grey[300],
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  AppDateUtils.formatDate(date),
                  style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppColors.textSecondary.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    date.isAfter(DateTime(today.year, today.month, today.day))
                        ? 'Ngày tương lai'
                        : 'Vắng mặt',
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),
            // Hiện trạng thái nghỉ phép nếu đã có đơn, ngược lại hiện button đăng ký
            if (existingLeave != null)
              _buildLeaveStatus(theme, existingLeave)
            else
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton.icon(
                  onPressed: () {
                    Navigator.pop(ctx);
                    _openLeaveRequest(date, null);
                  },
                  icon: const Icon(Icons.event_busy_outlined, size: 18),
                  label: const Text('Đăng ký nghỉ phép',
                    style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.calendarLeave,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  bool _isCorrectable(AttendanceModel record) {
    const correctableStatuses = {
      AppConstants.statusLate,
      AppConstants.statusEarlyLeave,
      AppConstants.statusLateEarlyLeave,
    };
    return correctableStatuses.contains(record.status);
  }

  void _openCorrectionRequest() async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => BlocProvider.value(
          value: context.read<CorrectionBloc>(),
          child: CorrectionRequestScreen(attendance: _selectedDayRecord!),
        ),
      ),
    );
    if (result == true) {
      _loadMonth(); // Reload history + corrections
    }
  }

  String _dateKey(DateTime date) => '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
}
