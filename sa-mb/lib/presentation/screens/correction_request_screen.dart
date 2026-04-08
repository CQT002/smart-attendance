import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/attendance_model.dart';
import '../blocs/correction/correction_bloc.dart';
import '../blocs/correction/correction_event.dart';
import '../blocs/correction/correction_state.dart';
import '../widgets/app_toast.dart';

/// Màn hình đăng ký bổ sung công — employee chọn từ lịch sử ngày bị trễ/về sớm
class CorrectionRequestScreen extends StatefulWidget {
  final AttendanceModel attendance;

  const CorrectionRequestScreen({super.key, required this.attendance});

  @override
  State<CorrectionRequestScreen> createState() => _CorrectionRequestScreenState();
}

class _CorrectionRequestScreenState extends State<CorrectionRequestScreen> {
  final _descriptionController = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  @override
  void dispose() {
    _descriptionController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final attendance = widget.attendance;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Đăng ký bổ sung công'),
        centerTitle: true,
      ),
      body: BlocListener<CorrectionBloc, CorrectionState>(
        listener: (context, state) {
          if (state is CorrectionCreateSuccess) {
            AppToast.show(context,
                message: 'Đã gửi yêu cầu bổ sung công thành công!',
                type: ToastType.success);
            Navigator.of(context).pop(true);
          } else if (state is CorrectionFailure) {
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
                // Thông tin ngày cần bù
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Thông tin ngày chấm công',
                            style: theme.textTheme.titleSmall
                                ?.copyWith(fontWeight: FontWeight.w700)),
                        const SizedBox(height: 12),
                        _infoRow('Ngày', AppDateUtils.formatDate(attendance.date)),
                        const SizedBox(height: 8),
                        _infoRow('Trạng thái', _statusLabel(attendance.status)),
                        const SizedBox(height: 8),
                        _infoRow(
                          'Check-in',
                          attendance.checkInTime != null
                              ? AppDateUtils.formatTime(attendance.checkInTime!)
                              : '--:--',
                        ),
                        const SizedBox(height: 8),
                        _infoRow(
                          'Check-out',
                          attendance.checkOutTime != null
                              ? AppDateUtils.formatTime(attendance.checkOutTime!)
                              : '--:--',
                        ),
                        const SizedBox(height: 8),
                        _infoRow(
                          'Giờ làm',
                          AppDateUtils.formatWorkHours(attendance.workHours),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 20),

                // Lý do
                Text('Lý do xin bổ sung công *',
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
                    hintText:
                        'Ví dụ: Do kẹt xe trên đường, em đã đến trễ 20 phút...',
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
                    color: AppColors.info.withOpacity(0.1),
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
                          'Bạn có tối đa 4 lần bổ sung công mỗi tháng. '
                          'Yêu cầu sẽ được Manager chi nhánh xem xét và duyệt.',
                          style: theme.textTheme.bodySmall
                              ?.copyWith(color: AppColors.info),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),

                // Submit button
                BlocBuilder<CorrectionBloc, CorrectionState>(
                  builder: (context, state) {
                    final isLoading = state is CorrectionLoading;
                    return SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: isLoading ? null : _submit,
                        child: isLoading
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Text('Gửi yêu cầu bổ sung công'),
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

  Widget _infoRow(String label, String value) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: const TextStyle(color: AppColors.textSecondary)),
        Text(value, style: const TextStyle(fontWeight: FontWeight.w600)),
      ],
    );
  }

  String _statusLabel(String status) {
    switch (status) {
      case 'late':
        return 'Đi trễ';
      case 'early_leave':
        return 'Về sớm';
      case 'late_early_leave':
        return 'Đi trễ - Về sớm';
      default:
        return status;
    }
  }

  void _submit() {
    if (!_formKey.currentState!.validate()) return;

    context.read<CorrectionBloc>().add(
          CorrectionCreateRequested(
            attendanceLogId: widget.attendance.id,
            description: _descriptionController.text.trim(),
          ),
        );
  }
}
