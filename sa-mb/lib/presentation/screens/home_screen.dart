import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../../data/models/attendance_model.dart';
import '../../data/models/user_model.dart';
import '../blocs/attendance/attendance_bloc.dart';
import '../blocs/attendance/attendance_event.dart';
import '../blocs/attendance/attendance_state.dart';
import '../blocs/auth/auth_bloc.dart';
import '../blocs/auth/auth_event.dart';
import '../blocs/auth/auth_state.dart';
import '../widgets/app_toast.dart';
import '../widgets/status_badge.dart';
import 'history_screen.dart';
import 'check_in_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _currentIndex = 0;

  @override
  void initState() {
    super.initState();
    context.read<AttendanceBloc>().add(AttendanceLoadToday());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
        index: _currentIndex,
        children: const [
          _HomeTab(),
          HistoryScreen(),
          _ProfileTab(),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _currentIndex,
        onDestinationSelected: (index) {
          setState(() => _currentIndex = index);
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.home_outlined),
            selectedIcon: Icon(Icons.home),
            label: 'Trang chủ',
          ),
          NavigationDestination(
            icon: Icon(Icons.history_outlined),
            selectedIcon: Icon(Icons.history),
            label: 'Lịch sử',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outlined),
            selectedIcon: Icon(Icons.person),
            label: 'Cá nhân',
          ),
        ],
      ),
    );
  }
}

class _HomeTab extends StatelessWidget {
  const _HomeTab();

  @override
  Widget build(BuildContext context) {
    final authState = context.watch<AuthBloc>().state;
    final user = authState is AuthAuthenticated ? authState.user : null;

    return SafeArea(
      child: RefreshIndicator(
        onRefresh: () async {
          context.read<AttendanceBloc>().add(AttendanceLoadToday());
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Greeting
              _buildGreeting(context, user),
              const SizedBox(height: 24),

              // Today status card
              _buildTodayCard(context),
              const SizedBox(height: 20),

              // Check-in/out buttons
              _buildActionButtons(context),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildGreeting(BuildContext context, UserModel? user) {
    final theme = Theme.of(context);
    final now = DateTime.now();
    String greeting;
    if (now.hour < 12) {
      greeting = 'Chào buổi sáng';
    } else if (now.hour < 18) {
      greeting = 'Chào buổi chiều';
    } else {
      greeting = 'Chào buổi tối';
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          greeting,
          style: theme.textTheme.bodyLarge?.copyWith(
            color: AppColors.textSecondary,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          user?.name ?? 'Nhân viên',
          style: theme.textTheme.headlineSmall?.copyWith(
            fontWeight: FontWeight.bold,
          ),
        ),
        if (user?.branch != null)
          Padding(
            padding: const EdgeInsets.only(top: 4),
            child: Row(
              children: [
                const Icon(Icons.business, size: 16, color: AppColors.textSecondary),
                const SizedBox(width: 4),
                Text(
                  user!.branch!.name,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
        const SizedBox(height: 4),
        Text(
          AppDateUtils.formatDate(DateTime.now()),
          style: theme.textTheme.bodyMedium?.copyWith(
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }

  Widget _buildTodayCard(BuildContext context) {
    final theme = Theme.of(context);

    return BlocBuilder<AttendanceBloc, AttendanceState>(
      builder: (context, state) {
        AttendanceModel? today;
        if (state is AttendanceTodayLoaded) {
          today = state.today;
        } else if (state is AttendanceCheckInSuccess) {
          today = state.attendance;
        } else if (state is AttendanceCheckOutSuccess) {
          today = state.attendance;
        }

        return Card(
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Hôm nay',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    if (today != null)
                      StatusBadge(
                        status: today.status,
                        label: today.statusDisplay,
                      ),
                  ],
                ),
                const SizedBox(height: 20),
                Row(
                  children: [
                    Expanded(
                      child: _buildTimeBlock(
                        context,
                        icon: Icons.login_rounded,
                        label: 'Check-in',
                        time: today?.checkInTime != null
                            ? AppDateUtils.formatTime(today!.checkInTime!)
                            : '--:--',
                        color: AppColors.success,
                      ),
                    ),
                    Container(
                      width: 1,
                      height: 60,
                      color: AppColors.divider,
                    ),
                    Expanded(
                      child: _buildTimeBlock(
                        context,
                        icon: Icons.logout_rounded,
                        label: 'Check-out',
                        time: today?.checkOutTime != null
                            ? AppDateUtils.formatTime(today!.checkOutTime!)
                            : '--:--',
                        color: AppColors.error,
                      ),
                    ),
                    Container(
                      width: 1,
                      height: 60,
                      color: AppColors.divider,
                    ),
                    Expanded(
                      child: _buildTimeBlock(
                        context,
                        icon: Icons.access_time_rounded,
                        label: 'Giờ làm',
                        time: today != null
                            ? AppDateUtils.formatWorkHours(today.workHours)
                            : '0h',
                        color: AppColors.info,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildTimeBlock(
    BuildContext context, {
    required IconData icon,
    required String label,
    required String time,
    required Color color,
  }) {
    return Column(
      children: [
        Icon(icon, color: color, size: 28),
        const SizedBox(height: 8),
        Text(
          time,
          style: Theme.of(context).textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.bold,
              ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: AppColors.textSecondary,
              ),
        ),
      ],
    );
  }

  Widget _buildActionButtons(BuildContext context) {
    return BlocConsumer<AttendanceBloc, AttendanceState>(
      listener: (context, state) {
        if (state is AttendanceCheckInSuccess) {
          AppToast.show(context,
              message: 'Check-in thành công!', type: ToastType.success);
        } else if (state is AttendanceCheckOutSuccess) {
          AppToast.show(context,
              message: 'Check-out thành công!', type: ToastType.success);
        } else if (state is AttendanceFailure) {
          AppToast.show(context, message: state.message);
        }
      },
      builder: (context, state) {
        final isLoading = state is AttendanceLoading;
        AttendanceModel? today;
        if (state is AttendanceTodayLoaded) today = state.today;
        if (state is AttendanceCheckInSuccess) today = state.attendance;
        if (state is AttendanceCheckOutSuccess) today = state.attendance;

        final hasCheckedIn = today?.hasCheckedIn ?? false;
        final hasCheckedOut = today?.hasCheckedOut ?? false;

        if (hasCheckedIn && hasCheckedOut) {
          return Card(
            color: AppColors.success.withOpacity(0.1),
            child: const Padding(
              padding: EdgeInsets.all(20),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.check_circle, color: AppColors.success),
                  SizedBox(width: 8),
                  Text(
                    'Bạn đã hoàn thành chấm công hôm nay',
                    style: TextStyle(
                      color: AppColors.success,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          );
        }

        return Column(
          children: [
            if (!hasCheckedIn)
              ElevatedButton.icon(
                onPressed: isLoading
                    ? null
                    : () => _showMethodPicker(context, isCheckIn: true),
                icon: isLoading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.login_rounded),
                label: Text(isLoading ? 'Đang xử lý...' : 'Check-in'),
              ),
            if (hasCheckedIn && !hasCheckedOut) ...[
              OutlinedButton.icon(
                onPressed: isLoading
                    ? null
                    : () => _showMethodPicker(
                          context,
                          isCheckIn: false,
                          attendanceId: today!.id,
                        ),
                icon: isLoading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.logout_rounded),
                label: Text(isLoading ? 'Đang xử lý...' : 'Check-out'),
              ),
            ],
          ],
        );
      },
    );
  }

  void _showMethodPicker(
    BuildContext context, {
    required bool isCheckIn,
    int? attendanceId,
  }) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => BlocProvider.value(
          value: context.read<AttendanceBloc>(),
          child: CheckInScreen(
            isCheckIn: isCheckIn,
            attendanceId: attendanceId,
          ),
        ),
      ),
    );
  }
}

class _ProfileTab extends StatelessWidget {
  const _ProfileTab();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return BlocBuilder<AuthBloc, AuthState>(
      builder: (context, state) {
        final user = state is AuthAuthenticated ? state.user : null;

        return SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(20),
            child: Column(
              children: [
                const SizedBox(height: 20),
                // Avatar
                CircleAvatar(
                  radius: 50,
                  backgroundColor: AppColors.primary.withOpacity(0.1),
                  child: Text(
                    user?.name.isNotEmpty == true
                        ? user!.name[0].toUpperCase()
                        : '?',
                    style: theme.textTheme.headlineLarge?.copyWith(
                      color: AppColors.primary,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                Text(
                  user?.name ?? '',
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  user?.position ?? '',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 32),

                // Info cards
                _buildInfoItem(context, Icons.badge_outlined, 'Mã NV', user?.employeeCode ?? ''),
                _buildInfoItem(context, Icons.email_outlined, 'Email', user?.email ?? ''),
                _buildInfoItem(context, Icons.phone_outlined, 'SĐT', user?.phone ?? ''),
                _buildInfoItem(context, Icons.business_outlined, 'Chi nhánh', user?.branch?.name ?? ''),
                _buildInfoItem(context, Icons.apartment_outlined, 'Phòng ban', user?.department ?? ''),
                const SizedBox(height: 32),

                // Logout button
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton.icon(
                    onPressed: () {
                      showDialog(
                        context: context,
                        builder: (ctx) => AlertDialog(
                          title: const Text('Đăng xuất'),
                          content: const Text('Bạn có chắc muốn đăng xuất?'),
                          actions: [
                            TextButton(
                              onPressed: () => Navigator.pop(ctx),
                              child: const Text('Huỷ'),
                            ),
                            TextButton(
                              onPressed: () {
                                Navigator.pop(ctx);
                                context.read<AuthBloc>().add(AuthLogoutRequested());
                              },
                              child: const Text('Đăng xuất'),
                            ),
                          ],
                        ),
                      );
                    },
                    icon: const Icon(Icons.logout, color: AppColors.error),
                    label: const Text(
                      'Đăng xuất',
                      style: TextStyle(color: AppColors.error),
                    ),
                    style: OutlinedButton.styleFrom(
                      side: const BorderSide(color: AppColors.error),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildInfoItem(
    BuildContext context,
    IconData icon,
    String label,
    String value,
  ) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Icon(icon, size: 22, color: AppColors.textSecondary),
          const SizedBox(width: 12),
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
                value.isEmpty ? '---' : value,
                style: Theme.of(context).textTheme.bodyLarge,
              ),
            ],
          ),
        ],
      ),
    );
  }
}
