import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_colors.dart';
import '../blocs/attendance/attendance_bloc.dart';
import '../blocs/attendance/attendance_event.dart';
import '../blocs/attendance/attendance_state.dart';
import '../widgets/app_toast.dart';

class CheckInScreen extends StatefulWidget {
  final bool isCheckIn;

  const CheckInScreen({
    super.key,
    required this.isCheckIn,
  });

  @override
  State<CheckInScreen> createState() => _CheckInScreenState();
}

class _CheckInScreenState extends State<CheckInScreen> {
  String _selectedMethod = 'wifi';

  void _performAction() {
    if (widget.isCheckIn) {
      context.read<AttendanceBloc>().add(
            AttendanceCheckInRequested(method: _selectedMethod),
          );
    } else {
      context.read<AttendanceBloc>().add(
            AttendanceCheckOutRequested(
              method: _selectedMethod,
            ),
          );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final title = widget.isCheckIn ? 'Check-in' : 'Check-out';

    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: BlocListener<AttendanceBloc, AttendanceState>(
        listener: (context, state) {
          if (state is AttendanceCheckInSuccess ||
              state is AttendanceCheckOutSuccess) {
            Navigator.of(context).pop();
          } else if (state is AttendanceFailure) {
            AppToast.show(context, message: state.message);
          }
        },
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Method selection
              Text(
                'Chọn phương thức chấm công',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 16),

              // WiFi option
              _buildMethodCard(
                icon: Icons.wifi,
                title: 'WiFi',
                subtitle: 'Quét thông tin WiFi văn phòng (SSID/BSSID)',
                value: 'wifi',
                color: AppColors.info,
              ),
              const SizedBox(height: 12),

              // GPS option
              _buildMethodCard(
                icon: Icons.gps_fixed,
                title: 'GPS',
                subtitle: 'Xác minh vị trí GPS (Geofencing)',
                value: 'gps',
                color: AppColors.success,
              ),
              const SizedBox(height: 32),

              // Security info
              Card(
                color: AppColors.warning.withOpacity(0.1),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      const Icon(Icons.shield_outlined, color: AppColors.warning),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Bảo mật',
                              style: theme.textTheme.titleSmall?.copyWith(
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              'Hệ thống sẽ kiểm tra VPN, Fake GPS và Device ID để chống gian lận.',
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: AppColors.textSecondary,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),

              const Spacer(),

              // Action button
              BlocBuilder<AttendanceBloc, AttendanceState>(
                builder: (context, state) {
                  final isLoading = state is AttendanceLoading;
                  return ElevatedButton.icon(
                    onPressed: isLoading ? null : _performAction,
                    icon: isLoading
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : Icon(
                            widget.isCheckIn
                                ? Icons.login_rounded
                                : Icons.logout_rounded,
                          ),
                    label: Text(isLoading ? 'Đang xử lý...' : title),
                    style: ElevatedButton.styleFrom(
                      backgroundColor:
                          widget.isCheckIn ? AppColors.primary : AppColors.secondary,
                      minimumSize: const Size(double.infinity, 56),
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

  Widget _buildMethodCard({
    required IconData icon,
    required String title,
    required String subtitle,
    required String value,
    required Color color,
  }) {
    final isSelected = _selectedMethod == value;

    return InkWell(
      onTap: () => setState(() => _selectedMethod = value),
      borderRadius: BorderRadius.circular(16),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isSelected ? color : AppColors.divider,
            width: isSelected ? 2 : 1,
          ),
          color: isSelected ? color.withOpacity(0.05) : null,
        ),
        child: Row(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: color.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, color: color),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: Theme.of(context).textTheme.titleSmall?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                ],
              ),
            ),
            Radio<String>(
              value: value,
              groupValue: _selectedMethod,
              onChanged: (v) => setState(() => _selectedMethod = v!),
              activeColor: color,
            ),
          ],
        ),
      ),
    );
  }
}
