import 'package:flutter/material.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/attendance_model.dart';
import 'status_badge.dart';

class AttendanceCard extends StatelessWidget {
  final AttendanceModel attendance;

  const AttendanceCard({super.key, required this.attendance});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  AppDateUtils.formatDate(attendance.date),
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
                ),
                StatusBadge(
                  status: attendance.status,
                  label: attendance.statusDisplay,
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                _buildTimeInfo(
                  context,
                  icon: Icons.login_rounded,
                  label: 'Vao',
                  time: attendance.checkInTime != null
                      ? AppDateUtils.formatTime(attendance.checkInTime!)
                      : '--:--',
                  method: attendance.checkInMethod,
                  color: AppColors.success,
                ),
                const SizedBox(width: 24),
                _buildTimeInfo(
                  context,
                  icon: Icons.logout_rounded,
                  label: 'Ra',
                  time: attendance.checkOutTime != null
                      ? AppDateUtils.formatTime(attendance.checkOutTime!)
                      : '--:--',
                  method: attendance.checkOutMethod,
                  color: AppColors.error,
                ),
                const Spacer(),
                if (attendance.workHours > 0)
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        AppDateUtils.formatWorkHours(attendance.workHours),
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                          color: AppColors.primary,
                        ),
                      ),
                      Text(
                        'Gio lam',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTimeInfo(
    BuildContext context, {
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
            Text(
              label,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: AppColors.textSecondary,
                  ),
            ),
            Text(
              time,
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
            ),
            if (method != null)
              Row(
                children: [
                  Icon(
                    method == 'wifi' ? Icons.wifi : Icons.gps_fixed,
                    size: 12,
                    color: AppColors.textSecondary,
                  ),
                  const SizedBox(width: 2),
                  Text(
                    method == 'wifi' ? 'WiFi' : 'GPS',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          fontSize: 10,
                          color: AppColors.textSecondary,
                        ),
                  ),
                ],
              ),
          ],
        ),
      ],
    );
  }
}
