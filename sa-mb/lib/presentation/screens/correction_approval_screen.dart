import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/correction_model.dart';
import '../blocs/correction/correction_bloc.dart';
import '../blocs/correction/correction_event.dart';
import '../blocs/correction/correction_state.dart';
import '../widgets/app_toast.dart';

/// Màn hình duyệt bù công dành cho Manager
class CorrectionApprovalScreen extends StatefulWidget {
  const CorrectionApprovalScreen({super.key});

  @override
  State<CorrectionApprovalScreen> createState() => _CorrectionApprovalScreenState();
}

class _CorrectionApprovalScreenState extends State<CorrectionApprovalScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  String _currentFilter = 'pending';

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _tabController.addListener(() {
      if (!_tabController.indexIsChanging) {
        final filters = ['pending', 'approved', 'rejected'];
        setState(() => _currentFilter = filters[_tabController.index]);
        _loadList();
      }
    });
    _loadList();
  }

  void _loadList() {
    context.read<CorrectionBloc>().add(
          CorrectionLoadAdminList(status: _currentFilter),
        );
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Duyệt bù công'),
        centerTitle: true,
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Chờ duyệt'),
            Tab(text: 'Đã duyệt'),
            Tab(text: 'Từ chối'),
          ],
        ),
      ),
      body: BlocConsumer<CorrectionBloc, CorrectionState>(
        listener: (context, state) {
          if (state is CorrectionProcessSuccess) {
            AppToast.show(context,
                message: state.message, type: ToastType.success);
            _loadList();
          } else if (state is CorrectionFailure) {
            AppToast.show(context, message: state.message);
          }
        },
        builder: (context, state) {
          if (state is CorrectionLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is CorrectionAdminListLoaded) {
            if (state.corrections.isEmpty) {
              return Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.inbox_outlined,
                        size: 64, color: AppColors.textSecondary.withValues(alpha: 0.5)),
                    const SizedBox(height: 16),
                    Text(
                      'Không có yêu cầu nào',
                      style: theme.textTheme.bodyLarge
                          ?.copyWith(color: AppColors.textSecondary),
                    ),
                  ],
                ),
              );
            }

            return RefreshIndicator(
              onRefresh: () async => _loadList(),
              child: ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: state.corrections.length,
                itemBuilder: (context, index) {
                  return _buildCorrectionCard(
                      theme, state.corrections[index]);
                },
              ),
            );
          }

          return const SizedBox.shrink();
        },
      ),
    );
  }

  Widget _buildCorrectionCard(ThemeData theme, CorrectionModel correction) {
    final attendance = correction.attendanceLog;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header: tên nhân viên + trạng thái
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: Text(
                    correction.user?.name ?? 'Nhân viên #${correction.userId}',
                    style: theme.textTheme.titleSmall
                        ?.copyWith(fontWeight: FontWeight.w700),
                  ),
                ),
                _statusChip(correction.status),
              ],
            ),
            if (correction.user != null) ...[
              const SizedBox(height: 4),
              Text(
                '${correction.user!.employeeCode} - ${correction.user!.department}',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: AppColors.textSecondary),
              ),
            ],
            const Divider(height: 20),

            // Thông tin ngày
            _infoRow('Ngày chấm công',
                attendance != null ? AppDateUtils.formatDate(attendance.date) : '---'),
            const SizedBox(height: 6),
            _infoRow('Trạng thái gốc', correction.originalStatusDisplay),
            const SizedBox(height: 6),
            if (attendance != null) ...[
              _infoRow(
                'Check-in / Check-out',
                '${attendance.checkInTime != null ? AppDateUtils.formatTime(attendance.checkInTime!) : "--:--"}'
                ' → '
                '${attendance.checkOutTime != null ? AppDateUtils.formatTime(attendance.checkOutTime!) : "--:--"}',
              ),
              const SizedBox(height: 6),
            ],
            _infoRow('Ngày gửi', AppDateUtils.formatDateTime(correction.createdAt)),
            const Divider(height: 20),

            // Lý do
            Text('Lý do:',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: AppColors.textSecondary)),
            const SizedBox(height: 4),
            Text(correction.description,
                style: theme.textTheme.bodyMedium),

            // Audit log (nếu đã xử lý)
            if (correction.processedAt != null) ...[
              const Divider(height: 20),
              _infoRow(
                'Người duyệt',
                correction.processedBy?.name ?? '---',
              ),
              const SizedBox(height: 6),
              _infoRow(
                'Thời gian duyệt',
                AppDateUtils.formatDateTime(correction.processedAt!),
              ),
              if (correction.managerNote.isNotEmpty) ...[
                const SizedBox(height: 6),
                Text('Ghi chú manager:',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: AppColors.textSecondary)),
                const SizedBox(height: 4),
                Text(correction.managerNote,
                    style: theme.textTheme.bodyMedium
                        ?.copyWith(fontStyle: FontStyle.italic)),
              ],
            ],

            // Action buttons (chỉ cho pending)
            if (correction.isPending) ...[
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => _showNoteDialog(context, correction, false),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: AppColors.error,
                        side: const BorderSide(color: AppColors.error),
                      ),
                      child: const Text('Từ chối'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: () => _showNoteDialog(context, correction, true),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.success,
                      ),
                      child: const Text('Duyệt'),
                    ),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _statusChip(String status) {
    Color color;
    String label;
    switch (status) {
      case 'pending':
        color = AppColors.warning;
        label = 'Chờ duyệt';
        break;
      case 'approved':
        color = AppColors.success;
        label = 'Đã duyệt';
        break;
      case 'rejected':
        color = AppColors.error;
        label = 'Từ chối';
        break;
      default:
        color = AppColors.textSecondary;
        label = status;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        label,
        style: TextStyle(
            color: color, fontSize: 12, fontWeight: FontWeight.w600),
      ),
    );
  }

  Widget _infoRow(String label, String value) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label,
            style: const TextStyle(
                color: AppColors.textSecondary, fontSize: 13)),
        Flexible(
          child: Text(value,
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
              textAlign: TextAlign.end),
        ),
      ],
    );
  }

  void _showNoteDialog(
      BuildContext context, CorrectionModel correction, bool isApprove) {
    final noteController = TextEditingController();
    final action = isApprove ? 'Duyệt' : 'Từ chối';

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('$action yêu cầu bù công'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Nhân viên: ${correction.user?.name ?? ""}',
              style: const TextStyle(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: noteController,
              maxLines: 3,
              decoration: InputDecoration(
                labelText: 'Ghi chú (không bắt buộc)',
                hintText: isApprove
                    ? 'Ví dụ: Đã xác nhận với phòng HC...'
                    : 'Ví dụ: Lý do không hợp lệ...',
                border: const OutlineInputBorder(),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Huỷ'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(ctx);
              if (isApprove) {
                context.read<CorrectionBloc>().add(
                      CorrectionApproveRequested(
                        correctionId: correction.id,
                        managerNote: noteController.text.trim(),
                      ),
                    );
              } else {
                context.read<CorrectionBloc>().add(
                      CorrectionRejectRequested(
                        correctionId: correction.id,
                        managerNote: noteController.text.trim(),
                      ),
                    );
              }
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: isApprove ? AppColors.success : AppColors.error,
            ),
            child: Text(action),
          ),
        ],
      ),
    );
  }
}
