import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import '../blocs/overtime/overtime_bloc.dart';
import '../blocs/overtime/overtime_event.dart';
import '../blocs/overtime/overtime_state.dart';
import '../widgets/app_toast.dart';

/// Màn hình chấm công tăng ca (Check-in/Check-out OT)
class OvertimeScreen extends StatefulWidget {
  const OvertimeScreen({super.key});

  @override
  State<OvertimeScreen> createState() => _OvertimeScreenState();
}

class _OvertimeScreenState extends State<OvertimeScreen> {
  @override
  void initState() {
    super.initState();
    context.read<OvertimeBloc>().add(OvertimeLoadToday());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chấm công tăng ca')),
      body: BlocConsumer<OvertimeBloc, OvertimeState>(
        listener: (context, state) {
          if (state is OvertimeCheckInSuccess) {
            AppToast.show(context, message: 'Check-in tăng ca thành công', type: ToastType.success);
            context.read<OvertimeBloc>().add(OvertimeLoadToday());
          } else if (state is OvertimeCheckOutSuccess) {
            AppToast.show(context, message: 'Check-out tăng ca thành công', type: ToastType.success);
            context.read<OvertimeBloc>().add(OvertimeLoadToday());
          } else if (state is OvertimeFailure) {
            AppToast.show(context, message: state.message, type: ToastType.error);
          }
        },
        builder: (context, state) {
          return SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // Lưu ý quy định
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.orange.shade50,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.orange.shade200),
                  ),
                  child: const Row(
                    children: [
                      Icon(Icons.info_outline, color: Colors.orange, size: 20),
                      SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'Lưu ý: Giờ tăng ca chỉ bắt đầu tính từ 18:00 đến 22:00 theo quy định',
                          style: TextStyle(fontSize: 13, color: Colors.orange),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),

                // Trạng thái OT hôm nay
                if (state is OvertimeTodayLoaded) ...[
                  _buildTodayStatus(state),
                ] else if (state is OvertimeLoading) ...[
                  const Center(child: CircularProgressIndicator()),
                ] else ...[
                  _buildNoOtToday(),
                ],

                const SizedBox(height: 24),

                // Nút Check-in / Check-out
                _buildActionButtons(state),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildTodayStatus(OvertimeTodayLoaded state) {
    final ot = state.overtime;
    if (ot == null) return _buildNoOtToday();

    final timeFormat = DateFormat('HH:mm');
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.schedule, color: Colors.blue),
                const SizedBox(width: 8),
                const Text('Tăng ca hôm nay',
                    style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                const Spacer(),
                _buildStatusChip(ot.status),
              ],
            ),
            const Divider(),
            if (ot.actualCheckin != null) ...[
              _buildInfoRow('Check-in thực tế', timeFormat.format(ot.actualCheckin!)),
            ],
            if (ot.actualCheckout != null) ...[
              _buildInfoRow('Check-out thực tế', timeFormat.format(ot.actualCheckout!)),
            ],
            if (ot.calculatedStart != null) ...[
              const Divider(),
              const Text('Giờ hệ thống tính (sau bo tròn):',
                  style: TextStyle(fontSize: 12, color: Colors.grey)),
              const SizedBox(height: 4),
              _buildInfoRow('Bắt đầu tính', timeFormat.format(ot.calculatedStart!)),
              if (ot.calculatedEnd != null)
                _buildInfoRow('Kết thúc tính', timeFormat.format(ot.calculatedEnd!)),
              _buildInfoRow('Tổng giờ OT', '${ot.totalHours} giờ'),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildNoOtToday() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Icon(Icons.nightlight_round, size: 48, color: Colors.grey.shade400),
            const SizedBox(height: 12),
            const Text('Chưa có tăng ca hôm nay',
                style: TextStyle(fontSize: 16, color: Colors.grey)),
            const SizedBox(height: 4),
            const Text('Bấm Check-in để bắt đầu ca tăng ca',
                style: TextStyle(fontSize: 13, color: Colors.grey)),
          ],
        ),
      ),
    );
  }

  Widget _buildActionButtons(OvertimeState state) {
    final isLoading = state is OvertimeLoading;
    final todayOt = state is OvertimeTodayLoaded ? state.overtime : null;
    final hasCheckedIn = todayOt?.isCheckedIn ?? false;
    final hasCheckedOut = todayOt?.isCheckedOut ?? false;

    if (!hasCheckedIn) {
      // Chưa check-in → hiện nút Check-in
      return ElevatedButton.icon(
        onPressed: isLoading
            ? null
            : () => context.read<OvertimeBloc>().add(OvertimeCheckInRequested()),
        icon: const Icon(Icons.login),
        label: const Text('Check-in Tăng ca'),
        style: ElevatedButton.styleFrom(
          padding: const EdgeInsets.symmetric(vertical: 14),
          backgroundColor: Colors.blue,
          foregroundColor: Colors.white,
        ),
      );
    } else if (!hasCheckedOut) {
      // Đã check-in, chưa check-out → hiện nút Check-out
      return ElevatedButton.icon(
        onPressed: isLoading
            ? null
            : () => context.read<OvertimeBloc>().add(
                  OvertimeCheckOutRequested(overtimeId: todayOt?.id),
                ),
        icon: const Icon(Icons.logout),
        label: const Text('Check-out Tăng ca'),
        style: ElevatedButton.styleFrom(
          padding: const EdgeInsets.symmetric(vertical: 14),
          backgroundColor: Colors.orange,
          foregroundColor: Colors.white,
        ),
      );
    } else {
      // Đã hoàn tất
      return Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.green.shade50,
          borderRadius: BorderRadius.circular(8),
        ),
        child: const Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.check_circle, color: Colors.green),
            SizedBox(width: 8),
            Text('Đã hoàn tất chấm công tăng ca hôm nay',
                style: TextStyle(color: Colors.green, fontWeight: FontWeight.w500)),
          ],
        ),
      );
    }
  }

  Widget _buildStatusChip(String status) {
    Color color;
    String label;
    switch (status) {
      case 'init':
        color = Colors.blue;
        label = 'Đang tăng ca';
        break;
      case 'approved':
        color = Colors.green;
        label = 'Đã duyệt';
        break;
      case 'rejected':
        color = Colors.red;
        label = 'Từ chối';
        break;
      default:
        color = Colors.orange;
        label = 'Chờ duyệt';
    }
    return Chip(
      label: Text(label, style: TextStyle(color: color, fontSize: 12)),
      backgroundColor: color.withOpacity(0.1),
      side: BorderSide.none,
      visualDensity: VisualDensity.compact,
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: Colors.grey)),
          Text(value, style: const TextStyle(fontWeight: FontWeight.w500)),
        ],
      ),
    );
  }
}
