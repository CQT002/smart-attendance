import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/attendance_model.dart';
import '../blocs/leave/leave_bloc.dart';
import '../blocs/leave/leave_event.dart';
import '../blocs/leave/leave_state.dart';
import '../widgets/app_toast.dart';

/// Màn hình đăng ký nghỉ phép
/// - Ngày quá khứ: hiển thị trạng thái thực tế, auto-detect loại nghỉ
/// - Ngày hiện tại/tương lai: cho chọn khung giờ (full_day, half_day_morning, half_day_afternoon)
class LeaveRequestScreen extends StatefulWidget {
  final DateTime selectedDate;
  final AttendanceModel? attendance; // null nếu là ngày tương lai

  const LeaveRequestScreen({
    super.key,
    required this.selectedDate,
    this.attendance,
  });

  @override
  State<LeaveRequestScreen> createState() => _LeaveRequestScreenState();
}

class _LeaveRequestScreenState extends State<LeaveRequestScreen> {
  final _descriptionController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  String _selectedLeaveType = 'full_day';

  bool get _isPastDate {
    final today = DateTime.now();
    final todayDate = DateTime(today.year, today.month, today.day);
    return widget.selectedDate.isBefore(todayDate);
  }

  @override
  void initState() {
    super.initState();
    if (_isPastDate && widget.attendance != null) {
      // Auto-detect leave type for past dates
      if (widget.attendance!.status == 'half_day') {
        // Nếu đã check-in buổi sáng → nghỉ buổi chiều
        if (widget.attendance!.checkInTime != null &&
            widget.attendance!.checkInTime!.hour < 12) {
          _selectedLeaveType = 'half_day_afternoon';
        } else {
          _selectedLeaveType = 'half_day_morning';
        }
      } else {
        _selectedLeaveType = 'full_day';
      }
    }
  }

  @override
  void dispose() {
    _descriptionController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Đăng ký nghỉ phép'),
        centerTitle: true,
      ),
      body: BlocListener<LeaveBloc, LeaveState>(
        listener: (context, state) {
          if (state is LeaveCreateSuccess) {
            AppToast.show(context,
                message: 'Đã gửi yêu cầu nghỉ phép thành công!',
                type: ToastType.success);
            Navigator.of(context).pop(true);
          } else if (state is LeaveFailure) {
            AppToast.show(context, message: state.message);
          }
        },
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Thông tin ngày nghỉ phép
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Thông tin ngày nghỉ',
                            style: theme.textTheme.titleSmall
                                ?.copyWith(fontWeight: FontWeight.w700)),
                        const SizedBox(height: 12),
                        _infoRow('Ngày',
                            AppDateUtils.formatDate(widget.selectedDate)),
                        if (_isPastDate && widget.attendance != null) ...[
                          const SizedBox(height: 8),
                          _infoRow('Trạng thái thực tế',
                              widget.attendance!.statusDisplay),
                          if (widget.attendance!.checkInTime != null) ...[
                            const SizedBox(height: 8),
                            _infoRow(
                              'Check-in',
                              AppDateUtils.formatTime(
                                  widget.attendance!.checkInTime!),
                            ),
                          ],
                          if (widget.attendance!.checkOutTime != null) ...[
                            const SizedBox(height: 8),
                            _infoRow(
                              'Check-out',
                              AppDateUtils.formatTime(
                                  widget.attendance!.checkOutTime!),
                            ),
                          ],
                        ],
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 20),

                // Loại nghỉ phép (chỉ cho chọn với ngày hiện tại/tương lai)
                if (!_isPastDate) ...[
                  Text('Khung giờ nghỉ *',
                      style: theme.textTheme.titleSmall
                          ?.copyWith(fontWeight: FontWeight.w700)),
                  const SizedBox(height: 8),
                  _buildLeaveTypeSelector(theme),
                  const SizedBox(height: 20),
                ] else ...[
                  // Hiển thị loại nghỉ phép tự động cho ngày quá khứ
                  Card(
                    color: AppColors.info.withValues(alpha:0.05),
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Row(
                        children: [
                          const Icon(Icons.auto_fix_high,
                              color: AppColors.info, size: 20),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text('Loại nghỉ phép',
                                    style: theme.textTheme.bodySmall
                                        ?.copyWith(color: AppColors.textSecondary)),
                                Text(_leaveTypeLabel(_selectedLeaveType),
                                    style: theme.textTheme.titleSmall
                                        ?.copyWith(fontWeight: FontWeight.w600)),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),
                ],

                // Lý do
                Text('Lý do xin nghỉ phép *',
                    style: theme.textTheme.titleSmall
                        ?.copyWith(fontWeight: FontWeight.w700)),
                const SizedBox(height: 8),
                TextFormField(
                  controller: _descriptionController,
                  maxLines: 4,
                  keyboardType: TextInputType.multiline,
                  textInputAction: TextInputAction.newline,
                  autocorrect: false,
                  enableSuggestions: true,
                  decoration: const InputDecoration(
                    hintText: 'Ví dụ: Nghỉ phép do việc gia đình...',
                    border: OutlineInputBorder(),
                  ),
                  validator: (value) {
                    if (value == null || value.trim().isEmpty) {
                      return 'Vui lòng nhập lý do';
                    }
                    if (value.trim().length < 10) {
                      return 'Lý do phải có ít nhất 10 ký tự';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 12),

                // Lưu ý
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppColors.info.withValues(alpha:0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Icon(Icons.info_outline,
                          color: AppColors.info, size: 20),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'Yêu cầu nghỉ phép sẽ được Manager chi nhánh xem xét và duyệt.',
                          style: theme.textTheme.bodySmall
                              ?.copyWith(color: AppColors.info),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),

                // Submit button
                BlocBuilder<LeaveBloc, LeaveState>(
                  builder: (context, state) {
                    final isLoading = state is LeaveLoading;
                    return SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: isLoading ? null : _submit,
                        child: isLoading
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child:
                                    CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Text('Gửi yêu cầu nghỉ phép'),
                      ),
                    );
                  },
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildLeaveTypeSelector(ThemeData theme) {
    final options = [
      ('full_day', 'Cả ngày', '08:00 - 17:00', Icons.wb_sunny),
      ('half_day_morning', 'Buổi sáng', '08:00 - 12:00', Icons.wb_twilight),
      ('half_day_afternoon', 'Buổi chiều', '13:00 - 17:00', Icons.nights_stay),
    ];

    return Column(
      children: options.map((option) {
        final isSelected = _selectedLeaveType == option.$1;
        return Padding(
          padding: const EdgeInsets.only(bottom: 8),
          child: InkWell(
            onTap: () => setState(() => _selectedLeaveType = option.$1),
            borderRadius: BorderRadius.circular(12),
            child: Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: isSelected ? AppColors.primary : Colors.grey[300]!,
                  width: isSelected ? 2 : 1,
                ),
                color: isSelected
                    ? AppColors.primary.withValues(alpha:0.05)
                    : null,
              ),
              child: Row(
                children: [
                  Icon(option.$4,
                      color: isSelected
                          ? AppColors.primary
                          : AppColors.textSecondary,
                      size: 24),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(option.$2,
                            style: theme.textTheme.titleSmall?.copyWith(
                                fontWeight: FontWeight.w600,
                                color: isSelected
                                    ? AppColors.primary
                                    : null)),
                        Text(option.$3,
                            style: theme.textTheme.bodySmall?.copyWith(
                                color: AppColors.textSecondary)),
                      ],
                    ),
                  ),
                  if (isSelected)
                    const Icon(Icons.check_circle,
                        color: AppColors.primary, size: 24),
                ],
              ),
            ),
          ),
        );
      }).toList(),
    );
  }

  Widget _infoRow(String label, String value) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: const TextStyle(color: AppColors.textSecondary)),
        Text(value, style: const TextStyle(fontWeight: FontWeight.w600)),
      ],
    );
  }

  String _leaveTypeLabel(String type) {
    switch (type) {
      case 'full_day':
        return 'Cả ngày (08:00 - 17:00)';
      case 'half_day_morning':
        return 'Buổi sáng (08:00 - 12:00)';
      case 'half_day_afternoon':
        return 'Buổi chiều (13:00 - 17:00)';
      default:
        return type;
    }
  }

  void _submit() {
    if (!_formKey.currentState!.validate()) return;

    final dateStr =
        '${widget.selectedDate.year}-${widget.selectedDate.month.toString().padLeft(2, '0')}-${widget.selectedDate.day.toString().padLeft(2, '0')}';

    context.read<LeaveBloc>().add(
          LeaveCreateRequested(
            leaveDate: dateStr,
            leaveType: _selectedLeaveType,
            description: _descriptionController.text.trim(),
          ),
        );
  }
}
