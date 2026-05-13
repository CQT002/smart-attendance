import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/constants/app_constants.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/attendance_model.dart';
import '../../data/models/correction_model.dart';
import '../../data/models/holiday_model.dart';
import '../../data/models/leave_model.dart';
import '../../data/models/overtime_model.dart';
import '../../domain/repositories/correction_repository.dart';
import '../../domain/repositories/holiday_repository.dart';
import '../../domain/repositories/leave_repository.dart';
import '../../domain/repositories/overtime_repository.dart';
import '../blocs/attendance/attendance_bloc.dart';
import '../blocs/attendance/attendance_event.dart';
import '../blocs/attendance/attendance_state.dart';
import '../blocs/correction/correction_bloc.dart';
import '../blocs/correction/correction_event.dart';
import '../blocs/correction/correction_state.dart';
import '../blocs/leave/leave_bloc.dart';
import '../widgets/app_toast.dart';
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
  Map<String, OvertimeModel> _overtimeByDate = {}; // "yyyy-MM-dd" → overtime
  Map<String, HolidayModel> _holidaysByDate = {}; // "yyyy-MM-dd" → holiday
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

  Future<void> _loadMonth() async {
    final from = AppDateUtils.startOfMonth(_currentMonth);
    final to = AppDateUtils.endOfMonth(_currentMonth);
    context.read<AttendanceBloc>().add(
          AttendanceLoadHistory(from: from, to: to),
        );
    setState(() {
      _selectedDayRecord = null;
      _selectedDay = null;
    });
    await Future.wait([
      _loadCorrections(),
      _loadLeaves(),
      _loadOvertime(),
      _loadHolidays(),
    ]);
  }

  Future<void> _loadCorrections() async {
    try {
      final repo = context.read<CorrectionRepository>();
      final corrections = await repo.getMyCorrections(limit: 100);
      setState(() {
        _correctionsByLogId = {
          for (final c in corrections)
            if (c.attendanceLogId != null) c.attendanceLogId!: c,
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

  Future<void> _loadOvertime() async {
    try {
      final repo = context.read<OvertimeRepository>();
      final overtimes = await repo.getMyList(limit: 100);
      setState(() {
        _overtimeByDate = {
          for (final ot in overtimes)
            _dateKey(ot.date): ot,
        };
      });
    } catch (_) {
      // Fail silently — overtime info is optional
    }
  }

  Future<void> _loadHolidays() async {
    try {
      final repo = context.read<HolidayRepository>();
      // Load theo phạm vi năm của tháng đang xem để bao phủ navigation qua lại
      final year = _currentMonth.year;
      final from = '${year.toString().padLeft(4, '0')}-01-01';
      final to = '${year.toString().padLeft(4, '0')}-12-31';
      final holidays = await repo.getCalendar(dateFrom: from, dateTo: to);
      if (!mounted) return;
      setState(() {
        _holidaysByDate = {for (final h in holidays) h.dateKey: h};
      });
    } catch (_) {
      // Fail silently — hiển thị calendar không có badge ngày lễ
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
      // Kiểm tra có OT record cho ngày này không (dù không có attendance log)
      final otRecord = _overtimeByDate[_dateKey(date)];
      if (otRecord != null) {
        _showOvertimeOnlySheet(date, otRecord);
      } else {
        // Ngày không có record — cho phép đăng ký nghỉ phép
        _showLeaveOnlySheet(date);
      }
    }
  }

  void _showDayDetail(AttendanceModel record) {
    final theme = Theme.of(context);
    final correctable = _isCorrectable(record);
    final existingCorrection = _correctionsByLogId[record.id];
    final existingLeave = _leavesByDate[_dateKey(record.date)];
    final existingOvertime = _overtimeByDate[_dateKey(record.date)];

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
                if (record.overtimeRequestId != null)
                  Container(
                    margin: const EdgeInsets.only(left: 6),
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: Colors.deepPurple.shade50,
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: const Text('OT', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.deepPurple)),
                  ),
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
                        label: const Text('Bổ sung công',
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
            // Thông tin tăng ca (nếu có)
            if (existingOvertime != null) ...[
              const SizedBox(height: 16),
              _buildOvertimeInfo(theme, existingOvertime, onClose: () => Navigator.pop(ctx)),
            ],
          ],
        ),
      ),
    );
  }

  void _openOvertimeCorrectionDialog(OvertimeModel ot) {
    final descController = TextEditingController();
    final formKey = GlobalKey<FormState>();
    final missing = (ot.isCheckedIn && !ot.isCheckedOut) ? 'check-out' : 'check-in';
    // Capture bloc reference before dialog (dialog has a different BuildContext)
    final correctionBloc = context.read<CorrectionBloc>();

    showDialog(
      context: context,
      builder: (ctx) => BlocProvider.value(
        value: correctionBloc,
        child: BlocListener<CorrectionBloc, CorrectionState>(
          listener: (context, state) {
            if (state is CorrectionCreateSuccess) {
              Navigator.of(ctx).pop();
              AppToast.show(context,
                  message: 'Đã gửi yêu cầu bổ sung công tăng ca!',
                  type: ToastType.success);
              _loadMonth();
            } else if (state is CorrectionFailure) {
              AppToast.show(context, message: state.message);
            }
          },
          child: AlertDialog(
            title: Row(
              children: [
                const Expanded(child: Text('Bổ sung công tăng ca')),
                GestureDetector(
                  onTap: () => Navigator.pop(ctx),
                  child: const Icon(Icons.close, size: 22, color: Colors.grey),
                ),
              ],
            ),
            content: Form(
              key: formKey,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Thiếu $missing tăng ca ngày ${AppDateUtils.formatDate(ot.date)}',
                      style: const TextStyle(fontSize: 13, color: AppColors.textSecondary)),
                  if (ot.actualCheckin != null)
                    Text('Check-in: ${AppDateUtils.formatTime(ot.actualCheckin!)}',
                        style: const TextStyle(fontSize: 13)),
                  if (ot.actualCheckout != null)
                    Text('Check-out: ${AppDateUtils.formatTime(ot.actualCheckout!)}',
                        style: const TextStyle(fontSize: 13)),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: descController,
                    maxLines: 3,
                    decoration: InputDecoration(
                      labelText: 'Lý do *',
                      hintText: 'Ví dụ: Quên $missing tăng ca...',
                      border: const OutlineInputBorder(),
                    ),
                    validator: (v) {
                      if (v == null || v.trim().isEmpty) return 'Vui lòng nhập lý do';
                      if (v.trim().length < 10) return 'Lý do phải có ít nhất 10 ký tự';
                      return null;
                    },
                  ),
                ],
              ),
            ),
            actionsPadding: const EdgeInsets.fromLTRB(24, 0, 24, 16),
            actions: [
              BlocBuilder<CorrectionBloc, CorrectionState>(
                builder: (context, state) {
                  final isLoading = state is CorrectionLoading;
                  return SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: isLoading
                          ? null
                          : () {
                              if (!formKey.currentState!.validate()) return;
                              correctionBloc.add(
                                    CorrectionCreateOvertimeRequested(
                                      overtimeRequestId: ot.id,
                                      description: descController.text.trim(),
                                    ),
                                  );
                            },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.deepPurple,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                      ),
                      child: isLoading
                          ? const SizedBox(width: 20, height: 20,
                              child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                          : const Text('Gửi yêu cầu', style: TextStyle(fontWeight: FontWeight.w600)),
                    ),
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildOvertimeInfo(ThemeData theme, OvertimeModel ot, {VoidCallback? onClose}) {
    final checkinStr = ot.actualCheckin != null
        ? AppDateUtils.formatTime(ot.actualCheckin!)
        : '--:--';
    final checkoutStr = ot.actualCheckout != null
        ? AppDateUtils.formatTime(ot.actualCheckout!)
        : '--:--';

    final needsCorrection = ot.isInit;
    final statusLabel = ot.statusDisplay;

    Color statusColor;
    switch (ot.status) {
      case 'approved':
        statusColor = AppColors.success;
        break;
      case 'rejected':
        statusColor = AppColors.error;
        break;
      case 'init':
        statusColor = Colors.blue;
        break;
      default:
        statusColor = AppColors.warning;
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.deepPurple.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.deepPurple.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.nightlight_round, size: 16, color: Colors.deepPurple),
              const SizedBox(width: 6),
              Text('Tăng ca', style: theme.textTheme.bodyMedium?.copyWith(
                fontWeight: FontWeight.w700, color: Colors.deepPurple)),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: statusColor.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(statusLabel, style: TextStyle(
                  fontSize: 11, fontWeight: FontWeight.w600, color: statusColor)),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Text('Check-in: $checkinStr', style: theme.textTheme.bodySmall),
              const SizedBox(width: 16),
              Text('Check-out: $checkoutStr', style: theme.textTheme.bodySmall),
              if (ot.totalHours > 0) ...[
                const Spacer(),
                Text('${ot.totalHours}h', style: theme.textTheme.bodySmall?.copyWith(
                  fontWeight: FontWeight.bold, color: Colors.deepPurple)),
              ],
            ],
          ),
          if (needsCorrection) ...[
            const SizedBox(height: 8),
            Text(
              ot.isCheckedIn && !ot.isCheckedOut
                  ? 'Thiếu check-out — cần đăng ký bổ sung công tăng ca'
                  : 'Thiếu check-in — cần đăng ký bổ sung công tăng ca',
              style: TextStyle(fontSize: 11, color: Colors.orange.shade700, fontStyle: FontStyle.italic),
            ),
            const SizedBox(height: 8),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: () {
                  if (onClose != null) onClose();
                  _openOvertimeCorrectionDialog(ot);
                },
                icon: const Icon(Icons.edit_calendar_outlined, size: 16),
                label: const Text('Bổ sung công tăng ca',
                  style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.deepPurple,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 10),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                ),
              ),
            ),
          ],
        ],
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
        statusText = 'Đã duyệt bổ sung công';
        break;
      case 'rejected':
        statusColor = AppColors.error;
        statusIcon = Icons.cancel_outlined;
        statusText = 'Từ chối bổ sung công';
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

                  return RefreshIndicator(
                    onRefresh: _loadMonth,
                    child: SingleChildScrollView(
                      controller: _scrollController,
                      physics: const AlwaysScrollableScrollPhysics(),
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
    final color = _statusColorForRecord(record);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.4)),
      ),
      child: Text(
        record.statusDisplay,
        style: TextStyle(
          color: color,
          fontSize: 12,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  /// Xác định màu dựa trên record thực tế (thiếu check-in/out → cam)
  Color _statusColorForRecord(AttendanceModel record) {
    final now = DateTime.now();
    final recordDate = DateTime(record.date.year, record.date.month, record.date.day);
    final today = DateTime(now.year, now.month, now.day);
    final isPastDay = recordDate.isBefore(today);

    // Thiếu check-out: đã check-in, chưa check-out, ngày đã qua
    if (record.hasCheckedIn && !record.hasCheckedOut && isPastDay) {
      return AppColors.warning;
    }
    // Thiếu check-in: chưa check-in, đã check-out, ngày đã qua
    if (!record.hasCheckedIn && record.hasCheckedOut && isPastDay) {
      return AppColors.warning;
    }

    return _statusColor(record.status);
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
        final hasOvertime = _overtimeByDate.containsKey(key);

        cells.add(
          Expanded(
            child: GestureDetector(
              onTap: () => _selectDay(date, record),
              child: SizedBox(
                height: 44,
                width: 44,
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    Container(
                      height: 40,
                      width: 40,
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
                    if (hasOvertime)
                      Positioned(
                        bottom: 0,
                        child: Container(
                          width: 6, height: 6,
                          decoration: const BoxDecoration(
                            shape: BoxShape.circle,
                            color: Colors.deepPurple,
                          ),
                        ),
                      ),
                  ],
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
    // Ngày lễ tương lai (T2-T6): vẫn hiển thị màu lễ để user biết trước
    if (isFuture) {
      if (isHoliday && !isWeekend) return AppColors.calendarHoliday;
      return null;
    }
    // Ngày lễ trong quá khứ/hôm nay: có đi làm → xanh lá (đủ công), còn lại → tím (paid holiday)
    if (isHoliday) {
      if (record != null && record.hasCheckedIn) {
        return AppColors.calendarPresent;
      }
      return AppColors.calendarHoliday;
    }
    if (isWeekend && record == null) return AppColors.calendarDayOff;
    if (record == null) return AppColors.calendarAbsent;
    if (record.status == 'leave' || record.status == 'approved_leave') return AppColors.calendarLeave;
    if (record.hasCheckedIn && record.hasCheckedOut) {
      return record.workHours >= 8 ? AppColors.calendarPresent : AppColors.calendarIncomplete;
    }
    if (record.hasCheckedIn || record.hasCheckedOut) return AppColors.calendarIncomplete;
    return AppColors.calendarAbsent;
  }

  bool _isHoliday(DateTime date) => _holidaysByDate.containsKey(_dateKey(date));

  HolidayModel? _holidayOn(DateTime date) => _holidaysByDate[_dateKey(date)];

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
            _legendItem(theme, AppColors.calendarHoliday, 'Ngày lễ'),
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
    final today = DateTime(now.year, now.month, now.day);
    final daysInMonth = DateTime(_currentMonth.year, _currentMonth.month + 1, 0).day;
    final isCurrentMonth = _currentMonth.year == now.year && _currentMonth.month == now.month;
    final isFutureMonth = _currentMonth.isAfter(DateTime(now.year, now.month, 1));

    // Số ngày làm việc tính đến hôm nay (tháng tương lai → 0; tháng quá khứ → cả tháng)
    final maxDayForWork = isFutureMonth ? 0 : (isCurrentMonth ? now.day : daysInMonth);

    var workDays = 0;
    for (var d = 1; d <= maxDayForWork; d++) {
      final date = DateTime(_currentMonth.year, _currentMonth.month, d);
      if (date.weekday < 6) workDays++;
    }

    // Holiday: đếm tất cả ngày lễ trong tháng (kể cả ngày chưa tới) để hiển thị section
    var holidayDays = 0;
    for (var d = 1; d <= daysInMonth; d++) {
      final date = DateTime(_currentMonth.year, _currentMonth.month, d);
      if (_isHoliday(date)) holidayDays++;
    }
    // Paid holiday chỉ tính ngày lễ đã qua/hôm nay không có check-in
    var paidHolidayPast = 0;
    for (var d = 1; d <= daysInMonth; d++) {
      final date = DateTime(_currentMonth.year, _currentMonth.month, d);
      if (date.isAfter(today)) continue;
      if (!_isHoliday(date)) continue;
      final rec = records.where((r) => _dateKey(r.date) == _dateKey(date)).cast<AttendanceModel?>().firstWhere((_) => true, orElse: () => null);
      if (rec == null || !rec.hasCheckedIn) paidHolidayPast++;
    }

    // Nghỉ phép đã duyệt (leave, half_day_leave) tính là đủ công
    const leaveStatuses = {'leave', 'half_day_leave'};

    // Ngày lễ có đi làm (có checkin) — chỉ tính ngày đã qua/hôm nay
    final recordsOnHoliday = records.where((r) =>
        _isHoliday(r.date) && !r.date.isAfter(today) && r.hasCheckedIn).toList();
    final holidayRecordDates = recordsOnHoliday.map((r) => _dateKey(r.date)).toSet();

    final fullDays = records.where((r) =>
        leaveStatuses.contains(r.status) ||
        (r.hasCheckedIn && r.hasCheckedOut && r.workHours >= 8)).length + paidHolidayPast;

    final incompleteDays = records.where((r) =>
        !leaveStatuses.contains(r.status) &&
        (r.hasCheckedIn || r.hasCheckedOut) &&
        !(r.hasCheckedIn && r.hasCheckedOut && r.workHours >= 8)).length;

    final absentDays = (workDays - fullDays - incompleteDays).clamp(0, workDays);

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
                _summaryItem(theme, '$absentDays', 'Vắng', AppColors.calendarAbsent),
                _summaryItem(theme, '$workDays', 'Ngày làm', AppColors.primary),
              ],
            ),
            if (holidayDays > 0) ...[
              const Divider(height: 24),
              Row(
                children: [
                  _summaryItem(theme, '${holidayRecordDates.length}', 'Làm ngày lễ', AppColors.calendarHoliday),
                  _summaryItem(theme, '$paidHolidayPast', 'Nghỉ lễ hưởng lương', AppColors.info),
                ],
              ),
            ],
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

  void _showOvertimeOnlySheet(DateTime date, OvertimeModel ot) {
    final theme = Theme.of(context);
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
                Text(AppDateUtils.formatDate(date),
                  style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700)),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.deepPurple.shade50,
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: const Text('OT', style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.deepPurple)),
                ),
              ],
            ),
            const SizedBox(height: 16),
            _buildOvertimeInfo(theme, ot, onClose: () => Navigator.pop(ctx)),
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton(
                onPressed: () => Navigator.pop(ctx),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                ),
                child: const Text('Đóng'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showLeaveOnlySheet(DateTime date) {
    final theme = Theme.of(context);
    final today = DateTime.now();
    final isWeekend = date.weekday == 6 || date.weekday == 7;
    final holiday = _holidayOn(date);

    // Không hiện sheet cho cuối tuần (trừ khi là ngày lễ — vẫn show để user biết)
    if (isWeekend && holiday == null) return;

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
                Builder(builder: (_) {
                  final isFuture = date.isAfter(DateTime(today.year, today.month, today.day));
                  // Ngày tương lai hoặc ngày lễ không có chấm công: không hiện chip Vắng mặt
                  if (isFuture || holiday != null) return const SizedBox.shrink();
                  return Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: AppColors.error.withValues(alpha: 0.15),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: AppColors.error.withValues(alpha: 0.4)),
                    ),
                    child: const Text(
                      'Vắng mặt',
                      style: TextStyle(
                        color: AppColors.error,
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  );
                }),
              ],
            ),
            const SizedBox(height: 20),
            if (holiday != null) ...[
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: AppColors.calendarHoliday.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: AppColors.calendarHoliday.withValues(alpha: 0.3)),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.celebration, color: AppColors.calendarHoliday, size: 20),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            holiday.name,
                            style: theme.textTheme.bodyMedium?.copyWith(
                              fontWeight: FontWeight.w700,
                              color: AppColors.calendarHoliday,
                            ),
                          ),
                          Text(
                            '${holiday.typeDisplay}${holiday.isCompensated ? " · Nghỉ bù" : ""}',
                            style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),
            ],
            // Hiện trạng thái nghỉ phép nếu đã có đơn, ngược lại hiện button đăng ký
            // Ngày lễ không có chấm công: chỉ hiển thị thông tin ngày lễ, không cho đăng ký nghỉ phép
            if (existingLeave != null)
              _buildLeaveStatus(theme, existingLeave)
            else if (holiday == null)
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
    if (correctableStatuses.contains(record.status)) return true;

    // Trường hợp thiếu check-in hoặc check-out, ngày đã qua
    final now = DateTime.now();
    final recordDate = DateTime(record.date.year, record.date.month, record.date.day);
    final today = DateTime(now.year, now.month, now.day);
    if (recordDate.isBefore(today)) {
      if (record.hasCheckedIn && !record.hasCheckedOut) return true;
      if (!record.hasCheckedIn && record.hasCheckedOut) return true;
    }

    return false;
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
