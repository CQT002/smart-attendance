import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/approval_item_model.dart';
import '../../domain/repositories/correction_repository.dart';
import '../blocs/correction/correction_bloc.dart';
import '../blocs/correction/correction_event.dart';
import '../blocs/correction/correction_state.dart';
import '../blocs/leave/leave_bloc.dart';
import '../blocs/leave/leave_event.dart';
import '../blocs/leave/leave_state.dart';
import '../blocs/overtime/overtime_bloc.dart';
import '../blocs/overtime/overtime_event.dart';
import '../blocs/overtime/overtime_state.dart';
import '../widgets/app_toast.dart';

/// Màn hình duyệt chấm công tổng hợp (bổ sung công + nghỉ phép) dành cho Manager
/// Sử dụng API unified GET /admin/approvals để lấy data đã merge sẵn từ backend
class CorrectionApprovalScreen extends StatefulWidget {
  /// Khi [isActive] chuyển từ false → true, screen tự reload data
  final bool isActive;

  const CorrectionApprovalScreen({super.key, this.isActive = true});

  @override
  State<CorrectionApprovalScreen> createState() => _CorrectionApprovalScreenState();
}

class _CorrectionApprovalScreenState extends State<CorrectionApprovalScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  String _currentFilter = 'pending';

  List<ApprovalItemModel> _items = [];
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _tabController.addListener(() {
      if (!_tabController.indexIsChanging) {
        final filters = ['pending', 'approved', 'rejected'];
        setState(() => _currentFilter = filters[_tabController.index]);
        _loadData();
      }
    });
    _loadData();
  }

  @override
  void didUpdateWidget(covariant CorrectionApprovalScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isActive && !oldWidget.isActive) {
      _loadData();
    }
  }

  /// Load data từ API unified — 1 request duy nhất, backend đã merge + sort sẵn
  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    try {
      final repo = context.read<CorrectionRepository>();
      final items = await repo.getApprovals(status: _currentFilter, limit: 100);

      if (!mounted) return;
      setState(() {
        _items = items;
        _isLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _isLoading = false);
    }
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    final pendingCount = _items.where((i) => i.isPending).length;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Duyệt chấm công/Tăng ca'),
        centerTitle: true,
        actions: [
          if (_currentFilter == 'pending' && pendingCount > 0)
            TextButton.icon(
              onPressed: () => _showBatchApproveConfirm(context, pendingCount),
              icon: const Icon(Icons.done_all, size: 18),
              label: const Text('Duyệt tất cả'),
              style: TextButton.styleFrom(foregroundColor: AppColors.success),
            ),
        ],
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Chờ duyệt'),
            Tab(text: 'Đã duyệt'),
            Tab(text: 'Từ chối'),
          ],
        ),
      ),
      body: MultiBlocListener(
        listeners: [
          BlocListener<CorrectionBloc, CorrectionState>(
            listener: (context, state) {
              if (state is CorrectionProcessSuccess) {
                AppToast.show(context, message: state.message, type: ToastType.success);
                _loadData();
              } else if (state is CorrectionBatchApproveSuccess) {
                AppToast.show(context,
                    message: 'Đã duyệt ${state.approvedCount} yêu cầu bổ sung công',
                    type: ToastType.success);
                _loadData();
              } else if (state is CorrectionFailure) {
                AppToast.show(context, message: state.message);
              }
            },
          ),
          BlocListener<LeaveBloc, LeaveState>(
            listener: (context, state) {
              if (state is LeaveProcessSuccess) {
                AppToast.show(context, message: state.message, type: ToastType.success);
                _loadData();
              } else if (state is LeaveBatchApproveSuccess) {
                AppToast.show(context,
                    message: 'Đã duyệt ${state.approvedCount} yêu cầu nghỉ phép',
                    type: ToastType.success);
                _loadData();
              } else if (state is LeaveFailure) {
                AppToast.show(context, message: state.message);
              }
            },
          ),
          BlocListener<OvertimeBloc, OvertimeState>(
            listener: (context, state) {
              if (state is OvertimeProcessSuccess) {
                AppToast.show(context, message: state.message, type: ToastType.success);
                _loadData();
              } else if (state is OvertimeBatchApproveSuccess) {
                AppToast.show(context,
                    message: 'Đã duyệt ${state.approvedCount} yêu cầu tăng ca',
                    type: ToastType.success);
                _loadData();
              } else if (state is OvertimeFailure) {
                AppToast.show(context, message: state.message);
              }
            },
          ),
        ],
        child: _buildBody(theme),
      ),
    );
  }

  Widget _buildBody(ThemeData theme) {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_items.isEmpty) {
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
      onRefresh: () async => _loadData(),
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 100),
        itemCount: _items.length + 1, // +1 cho header count
        itemBuilder: (context, index) {
          if (index == 0) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Text(
                'Tổng ${_items.length} yêu cầu',
                style: theme.textTheme.bodySmall?.copyWith(
                  color: AppColors.textSecondary,
                  fontWeight: FontWeight.w500,
                ),
                textAlign: TextAlign.end,
              ),
            );
          }
          return _buildItemCard(theme, _items[index - 1]);
        },
      ),
    );
  }

  Widget _buildItemCard(ThemeData theme, ApprovalItemModel item) {
    if (item.isCorrection) {
      return _buildCorrectionCard(theme, item);
    } else if (item.isOvertime) {
      return _buildOvertimeCard(theme, item);
    } else {
      return _buildLeaveCard(theme, item);
    }
  }

  Widget _buildCorrectionCard(ThemeData theme, ApprovalItemModel item) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                _typeChip('Bổ sung công', AppColors.primary),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    item.userName.isNotEmpty ? item.userName : 'Nhân viên #${item.userId}',
                    style: theme.textTheme.titleSmall
                        ?.copyWith(fontWeight: FontWeight.w700),
                  ),
                ),
                _statusChip(item.status),
              ],
            ),
            if (item.employeeCode.isNotEmpty) ...[
              const SizedBox(height: 4),
              Text(
                '${item.employeeCode} - ${item.department}',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: AppColors.textSecondary),
              ),
            ],
            const Divider(height: 20),
            _infoRow('Ngày chấm công', item.date.isNotEmpty ? item.date : '---'),
            const SizedBox(height: 6),
            _infoRow('Trạng thái gốc', item.originalStatusDisplay),
            const SizedBox(height: 6),
            if (item.checkInTime != null || item.checkOutTime != null) ...[
              _infoRow(
                'Check-in / Check-out',
                '${item.checkInTime ?? "--:--"} → ${item.checkOutTime ?? "--:--"}',
              ),
              const SizedBox(height: 6),
            ],
            _infoRow('Ngày gửi', AppDateUtils.formatDateTime(item.createdAt)),
            const Divider(height: 20),
            Text('Lý do:',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: AppColors.textSecondary)),
            const SizedBox(height: 4),
            Text(item.description, style: theme.textTheme.bodyMedium),
            if (item.processedAt != null) ...[
              const Divider(height: 20),
              _infoRow('Người duyệt', item.processedByName.isNotEmpty ? item.processedByName : '---'),
              const SizedBox(height: 6),
              _infoRow('Thời gian duyệt',
                  AppDateUtils.formatDateTime(item.processedAt!)),
              if (item.managerNote.isNotEmpty) ...[
                const SizedBox(height: 6),
                Text('Ghi chú manager:',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: AppColors.textSecondary)),
                const SizedBox(height: 4),
                Text(item.managerNote,
                    style: theme.textTheme.bodyMedium
                        ?.copyWith(fontStyle: FontStyle.italic)),
              ],
            ],
            if (item.isPending) ...[
              const SizedBox(height: 16),
              _buildActionButtons(
                context,
                onApprove: () => _showNoteDialog(context, item, true),
                onReject: () => _showNoteDialog(context, item, false),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildLeaveCard(ThemeData theme, ApprovalItemModel item) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                _typeChip('Nghỉ phép', AppColors.calendarLeave),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    item.userName.isNotEmpty ? item.userName : 'Nhân viên #${item.userId}',
                    style: theme.textTheme.titleSmall
                        ?.copyWith(fontWeight: FontWeight.w700),
                  ),
                ),
                _statusChip(item.status),
              ],
            ),
            if (item.employeeCode.isNotEmpty) ...[
              const SizedBox(height: 4),
              Text(
                '${item.employeeCode} - ${item.department}',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: AppColors.textSecondary),
              ),
            ],
            const Divider(height: 20),
            _infoRow('Ngày nghỉ', item.date),
            const SizedBox(height: 6),
            _infoRow('Loại nghỉ phép', item.leaveTypeDisplay),
            const SizedBox(height: 6),
            _infoRow('Khung giờ', item.timeRangeDisplay),
            const SizedBox(height: 6),
            _infoRow('Ngày gửi', AppDateUtils.formatDateTime(item.createdAt)),
            if (item.detail.isNotEmpty && item.detail != item.leaveType) ...[
              const SizedBox(height: 6),
              _infoRow('Trạng thái gốc', item.originalStatusDisplay),
            ],
            const Divider(height: 20),
            Text('Lý do:',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: AppColors.textSecondary)),
            const SizedBox(height: 4),
            Text(item.description, style: theme.textTheme.bodyMedium),
            if (item.processedAt != null) ...[
              const Divider(height: 20),
              _infoRow('Người duyệt', item.processedByName.isNotEmpty ? item.processedByName : '---'),
              const SizedBox(height: 6),
              _infoRow('Thời gian duyệt',
                  AppDateUtils.formatDateTime(item.processedAt!)),
              if (item.managerNote.isNotEmpty) ...[
                const SizedBox(height: 6),
                Text('Ghi chú manager:',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: AppColors.textSecondary)),
                const SizedBox(height: 4),
                Text(item.managerNote,
                    style: theme.textTheme.bodyMedium
                        ?.copyWith(fontStyle: FontStyle.italic)),
              ],
            ],
            if (item.isPending) ...[
              const SizedBox(height: 16),
              _buildActionButtons(
                context,
                onApprove: () => _showNoteDialog(context, item, true),
                onReject: () => _showNoteDialog(context, item, false),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildOvertimeCard(ThemeData theme, ApprovalItemModel item) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                _typeChip('Tăng ca', Colors.deepPurple),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    item.userName.isNotEmpty ? item.userName : 'Nhân viên #${item.userId}',
                    style: theme.textTheme.titleSmall
                        ?.copyWith(fontWeight: FontWeight.w700),
                  ),
                ),
                _statusChip(item.status),
              ],
            ),
            if (item.employeeCode.isNotEmpty) ...[
              const SizedBox(height: 4),
              Text(
                '${item.employeeCode} - ${item.department}',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: AppColors.textSecondary),
              ),
            ],
            const Divider(height: 20),
            _infoRow('Ngày tăng ca', item.date),
            const SizedBox(height: 6),
            if (item.actualCheckin != null)
              ...[_infoRow('Check-in thực tế', item.actualCheckin!), const SizedBox(height: 6)],
            if (item.actualCheckout != null)
              ...[_infoRow('Check-out thực tế', item.actualCheckout!), const SizedBox(height: 6)],
            if (item.calculatedStart != null && item.calculatedEnd != null) ...[
              _infoRow('Giờ tính (bo tròn)', '${item.calculatedStart} - ${item.calculatedEnd}'),
              const SizedBox(height: 6),
            ],
            if (item.totalHours != null && item.totalHours! > 0)
              ...[_infoRow('Tổng giờ OT', '${item.totalHours}h'), const SizedBox(height: 6)],
            _infoRow('Trạng thái gốc', item.originalStatusDisplay),
            const SizedBox(height: 6),
            _infoRow('Ngày gửi', AppDateUtils.formatDateTime(item.createdAt)),
            const Divider(height: 20),
            Text('Lý do:',
                style: theme.textTheme.bodySmall
                    ?.copyWith(color: AppColors.textSecondary)),
            const SizedBox(height: 4),
            Text(item.description, style: theme.textTheme.bodyMedium),
            if (item.processedAt != null) ...[
              const Divider(height: 20),
              _infoRow('Người duyệt', item.processedByName.isNotEmpty ? item.processedByName : '---'),
              const SizedBox(height: 6),
              _infoRow('Thời gian duyệt',
                  AppDateUtils.formatDateTime(item.processedAt!)),
              if (item.managerNote.isNotEmpty) ...[
                const SizedBox(height: 6),
                Text('Ghi chú manager:',
                    style: theme.textTheme.bodySmall
                        ?.copyWith(color: AppColors.textSecondary)),
                const SizedBox(height: 4),
                Text(item.managerNote,
                    style: theme.textTheme.bodyMedium
                        ?.copyWith(fontStyle: FontStyle.italic)),
              ],
            ],
            if (item.isPending) ...[
              const SizedBox(height: 16),
              _buildActionButtons(
                context,
                onApprove: () => _showNoteDialog(context, item, true),
                onReject: () => _showNoteDialog(context, item, false),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildActionButtons(BuildContext context,
      {required VoidCallback onApprove, required VoidCallback onReject}) {
    return Row(
      children: [
        Expanded(
          child: OutlinedButton(
            onPressed: onReject,
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
            onPressed: onApprove,
            style: ElevatedButton.styleFrom(backgroundColor: AppColors.success),
            child: const Text('Duyệt'),
          ),
        ),
      ],
    );
  }

  Widget _typeChip(String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(
        label,
        style: TextStyle(color: color, fontSize: 11, fontWeight: FontWeight.w700),
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
        style: TextStyle(color: color, fontSize: 12, fontWeight: FontWeight.w600),
      ),
    );
  }

  Widget _infoRow(String label, String value) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label,
            style: const TextStyle(color: AppColors.textSecondary, fontSize: 13)),
        Flexible(
          child: Text(value,
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
              textAlign: TextAlign.end),
        ),
      ],
    );
  }

  void _showBatchApproveConfirm(BuildContext context, int pendingCount) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            const Expanded(child: Text('Xác nhận duyệt tất cả')),
            GestureDetector(
              onTap: () => Navigator.pop(ctx),
              child: const Icon(Icons.close, size: 22, color: Colors.grey),
            ),
          ],
        ),
        content: Text(
          'Bạn có chắc muốn duyệt tất cả $pendingCount yêu cầu đang chờ?\n\n'
          'Hành động này không thể hoàn tác.',
        ),
        actionsPadding: const EdgeInsets.fromLTRB(20, 0, 20, 16),
        actions: [
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: () {
                Navigator.pop(ctx);
                context.read<CorrectionBloc>().add(CorrectionBatchApproveRequested());
                context.read<LeaveBloc>().add(LeaveBatchApproveRequested());
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.success,
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
              ),
              child: const Text('Duyệt tất cả', style: TextStyle(fontWeight: FontWeight.w600)),
            ),
          ),
        ],
      ),
    );
  }

  void _showNoteDialog(
      BuildContext context, ApprovalItemModel item, bool isApprove) {
    final noteController = TextEditingController();
    final action = isApprove ? 'Duyệt' : 'Từ chối';
    final typeLabel = item.isCorrection ? 'bổ sung công' : (item.isOvertime ? 'tăng ca' : 'nghỉ phép');

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Row(
          children: [
            Expanded(child: Text('$action yêu cầu $typeLabel')),
            GestureDetector(
              onTap: () => Navigator.pop(ctx),
              child: const Icon(Icons.close, size: 22, color: Colors.grey),
            ),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Nhân viên: ${item.userName}',
              style: const TextStyle(fontWeight: FontWeight.w600),
            ),
            if (item.isLeave) ...[
              const SizedBox(height: 4),
              Text(
                'Ngày: ${item.date} - ${item.leaveTypeDisplay}',
                style: const TextStyle(color: AppColors.textSecondary),
              ),
            ],
            const SizedBox(height: 12),
            TextField(
              controller: noteController,
              maxLines: 3,
              decoration: InputDecoration(
                labelText: 'Ghi chú (không bắt buộc)',
                hintText: isApprove
                    ? 'Ví dụ: Đã xác nhận...'
                    : 'Ví dụ: Không đủ điều kiện...',
                border: const OutlineInputBorder(),
              ),
            ),
          ],
        ),
        actionsPadding: const EdgeInsets.fromLTRB(24, 0, 24, 16),
        actions: [
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: () {
                Navigator.pop(ctx);
                final note = noteController.text.trim();
                if (item.isCorrection) {
                  if (isApprove) {
                    context.read<CorrectionBloc>().add(
                          CorrectionApproveRequested(
                            correctionId: item.id,
                            managerNote: note,
                          ),
                        );
                  } else {
                    context.read<CorrectionBloc>().add(
                          CorrectionRejectRequested(
                            correctionId: item.id,
                            managerNote: note,
                          ),
                        );
                  }
                } else if (item.isOvertime) {
                  if (isApprove) {
                    context.read<OvertimeBloc>().add(
                          OvertimeApproveRequested(
                            overtimeId: item.id,
                            managerNote: note,
                          ),
                        );
                  } else {
                    context.read<OvertimeBloc>().add(
                          OvertimeRejectRequested(
                            overtimeId: item.id,
                            managerNote: note,
                          ),
                        );
                  }
                } else {
                  if (isApprove) {
                    context.read<LeaveBloc>().add(
                          LeaveApproveRequested(
                            leaveId: item.id,
                            managerNote: note,
                          ),
                        );
                  } else {
                    context.read<LeaveBloc>().add(
                          LeaveRejectRequested(
                            leaveId: item.id,
                            managerNote: note,
                          ),
                        );
                  }
                }
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: isApprove ? AppColors.success : AppColors.error,
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
              ),
              child: Text(action, style: const TextStyle(fontWeight: FontWeight.w600)),
            ),
          ),
        ],
      ),
    );
  }
}
