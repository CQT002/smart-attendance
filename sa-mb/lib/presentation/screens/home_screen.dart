import 'dart:async';
import 'package:flutter/material.dart';
import '../../../domain/repositories/attendance_repository.dart';
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

class _HomeTab extends StatefulWidget {
  const _HomeTab();

  @override
  State<_HomeTab> createState() => _HomeTabState();
}

class _HomeTabState extends State<_HomeTab> {
  Timer? _timer;
  DateTime _currentTime = DateTime.now();
  late Future<List<AttendanceModel>> _historyFuture;
  AttendanceModel? _lastKnownToday;

  @override
  void initState() {
    super.initState();
    _startTimer();
    _fetchHistory();
    _fetchToday();
  }

  void _fetchToday() async {
    final repo = context.read<AttendanceRepository>();
    final today = await repo.getTodayAttendance();
    if (mounted && today != null) {
      setState(() => _lastKnownToday = today);
    }
  }

  void _startTimer() {
    _timer = Timer.periodic(const Duration(minutes: 1), (timer) {
      if (mounted) {
        setState(() {
          _currentTime = DateTime.now();
        });
      }
    });
  }

  void _fetchHistory() {
    final repo = context.read<AttendanceRepository>();
    final from = AppDateUtils.startOfWeek(DateTime.now());
    final to = AppDateUtils.endOfDay(DateTime.now());
    _historyFuture = repo.getHistory(from: from, to: to, limit: 100);
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final authState = context.watch<AuthBloc>().state;
    final user = authState is AuthAuthenticated ? authState.user : null;

    return SafeArea(
      child: RefreshIndicator(
        onRefresh: () async {
          _fetchToday();
          _fetchHistory();
          setState(() {});
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildGreeting(context, user),
              const SizedBox(height: 24),
              _buildRealtimeCard(context),
              const SizedBox(height: 24),
              _buildWeeklyHistory(context),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildGreeting(BuildContext context, UserModel? user) {
    final theme = Theme.of(context);
    
    return Row(
      children: [
        CircleAvatar(
          radius: 28,
          backgroundColor: AppColors.primary.withOpacity(0.1),
          // Fallback to initial
          child: Text(
            user?.name.isNotEmpty == true ? user!.name[0].toUpperCase() : '?',
            style: theme.textTheme.headlineSmall?.copyWith(
              color: AppColors.primary,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                _getGreeting(),
                style: theme.textTheme.bodyLarge?.copyWith(
                  color: AppColors.textSecondary,
                  fontSize: 15,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                (user?.name ?? 'Nhân viên').toUpperCase(),
                style: theme.textTheme.titleLarge?.copyWith(
                  color: AppColors.textPrimary,
                  fontWeight: FontWeight.w800,
                  fontSize: 20,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                user?.branch?.name ?? 'Chi nhánh',
                style: theme.textTheme.bodyLarge?.copyWith(
                  color: AppColors.primary,
                  fontWeight: FontWeight.w600,
                  fontSize: 14,
                ),
              ),
            ],
          ),
        ),
        IconButton(
          onPressed: () {},
          icon: const Icon(Icons.notifications_outlined, size: 28),
        ),
      ],
    );
  }

  Widget _buildRealtimeCard(BuildContext context) {
    final theme = Theme.of(context);

    // Dùng _lastKnownToday từ _fetchToday() + listener reload sau check-in/out
    return BlocListener<AttendanceBloc, AttendanceState>(
      listener: (context, state) {
        if (state is AttendanceCheckInSuccess) {
          AppToast.show(context, message: 'Check-in thành công!', type: ToastType.success);
          _fetchToday(); // Reload data mới nhất từ API
          _fetchHistory();
        } else if (state is AttendanceCheckOutSuccess) {
          AppToast.show(context, message: 'Check-out thành công!', type: ToastType.success);
          _fetchToday(); // Reload data mới nhất từ API
          _fetchHistory();
        } else if (state is AttendanceFailure) {
          AppToast.show(context, message: state.message);
        }
      },
      child: Builder(
      builder: (context) {
        final isLoading = context.watch<AttendanceBloc>().state is AttendanceLoading;
        final today = _lastKnownToday;
        final hasCheckedIn = today?.hasCheckedIn ?? false;
        final hasCheckedOut = today?.hasCheckedOut ?? false;

        final checkInStr = today?.checkInTime != null ? AppDateUtils.formatTime(today!.checkInTime!) : '--:--';
        final checkOutStr = today?.checkOutTime != null ? AppDateUtils.formatTime(today!.checkOutTime!) : '--:--';

        // Calculate hours worked
        String workedStr = '0h 0p';
        if (hasCheckedIn && today!.checkInTime != null) {
          final endTime = today.checkOutTime ?? _currentTime;
          final diff = endTime.difference(today.checkInTime!);
          final h = diff.inHours;
          final m = diff.inMinutes.remainder(60);
          workedStr = '${h}h ${m}p';
        }

        return Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.05),
                blurRadius: 10,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: Column(
            children: [
              // Header of card
              Container(
                padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
                decoration: const BoxDecoration(
                  color: Color(0xFF005b8a), // Dark teal/blue matching image
                  borderRadius: BorderRadius.only(
                    topLeft: Radius.circular(16),
                    topRight: Radius.circular(16),
                  ),
                ),
                child: Row(
                  children: [
                    Icon(Icons.history, color: Colors.white.withOpacity(0.8), size: 18),
                    const SizedBox(width: 8),
                    Text(
                      '${AppDateUtils.formatDayName(DateTime.now())}, ${AppDateUtils.formatDate(DateTime.now())} Hôm nay',
                      style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600, fontSize: 13),
                    ),
                  ],
                ),
              ),
              
              // Big Time display
              Padding(
                padding: const EdgeInsets.all(20),
                child: Column(
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                      children: [
                        Text(
                          checkInStr,
                          style: const TextStyle(
                            fontSize: 32,
                            fontWeight: FontWeight.bold,
                            color: Colors.orange,
                          ),
                        ),
                        Text(
                          checkOutStr,
                          style: TextStyle(
                            fontSize: 32,
                            fontWeight: FontWeight.bold,
                            color: hasCheckedOut ? AppColors.primary : AppColors.textSecondary,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 20),

                    // Buttons — luôn enable, backend xử lý idempotent
                    Row(
                      children: [
                        Expanded(
                          child: ElevatedButton(
                            onPressed: isLoading
                                ? null
                                : () => _showMethodPicker(context, isCheckIn: true),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.orange,
                              disabledBackgroundColor: Colors.orange.withOpacity(0.5),
                              foregroundColor: Colors.white,
                              elevation: 0,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(30),
                              ),
                              padding: const EdgeInsets.symmetric(vertical: 14),
                            ),
                            child: isLoading
                                ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                                : const Text(
                                    'CHECK-IN',
                                    style: TextStyle(
                                      fontWeight: FontWeight.bold,
                                      color: Colors.white,
                                    ),
                                  ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: ElevatedButton(
                            onPressed: isLoading
                                ? null
                                : () => _showMethodPicker(context, isCheckIn: false),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: AppColors.primary,
                              disabledBackgroundColor: AppColors.divider.withOpacity(0.5),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(30),
                              ),
                              padding: const EdgeInsets.symmetric(vertical: 14),
                            ),
                            child: isLoading
                                ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
                                : const Text(
                                    'CHECK-OUT',
                                    style: TextStyle(fontWeight: FontWeight.bold),
                                  ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 24),
                    
                    // Times row
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          checkInStr,
                          style: const TextStyle(
                            color: Colors.orange,
                            fontWeight: FontWeight.bold,
                            fontSize: 16,
                          ),
                        ),
                        Text(
                          workedStr,
                          style: const TextStyle(
                            color: AppColors.textSecondary,
                            fontSize: 13,
                          ),
                        ),
                        Text(
                          hasCheckedOut ? checkOutStr : (hasCheckedIn ? AppDateUtils.formatTime(_currentTime) : '--:--'),
                          style: const TextStyle(
                            color: AppColors.primary,
                            fontWeight: FontWeight.bold,
                            fontSize: 16,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
      ),
    );
  }

  Widget _buildWeeklyHistory(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
           mainAxisAlignment: MainAxisAlignment.spaceBetween,
           children: [
             Text(
               'Tuần này (${AppDateUtils.formatDate(AppDateUtils.startOfWeek(DateTime.now()))} - ${AppDateUtils.formatDate(AppDateUtils.endOfWeek(DateTime.now()))})',
               style: const TextStyle(color: AppColors.textSecondary, fontWeight: FontWeight.w600, fontSize: 13),
             ),
             const Icon(Icons.keyboard_arrow_up, color: AppColors.textSecondary),
           ],
        ),
        const SizedBox(height: 16),
        FutureBuilder<List<AttendanceModel>>(
          future: _historyFuture,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const Center(child: Padding(padding: EdgeInsets.all(20), child: CircularProgressIndicator()));
            }
            if (snapshot.hasError) {
              return const Center(child: Padding(padding: EdgeInsets.all(20), child: Text('Không thể tải lịch sử')));
            }
            
            final list = snapshot.data ?? [];
            if (list.isEmpty) {
              return const Center(child: Padding(padding: EdgeInsets.all(20), child: Text('Chưa có lịch sử tuần này')));
            }

            // Exclude today from memory to not duplicate info, or just show it depending on preference. 
            // The mockup shows today's date in the history too. Let's just group or list them.
            return ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: list.length,
              separatorBuilder: (_, __) => const Divider(height: 32),
              itemBuilder: (context, index) {
                final item = list[index];
                final inStr = item.checkInTime != null ? AppDateUtils.formatTime(item.checkInTime!) : '--:--';
                final outStr = item.checkOutTime != null ? AppDateUtils.formatTime(item.checkOutTime!) : '--:--';
                
                return Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${AppDateUtils.formatDayName(item.date)}, ${AppDateUtils.formatDate(item.date)}',
                          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'Cả ngày - ${AppDateUtils.formatWorkHours(item.workHours)}',
                          style: const TextStyle(color: AppColors.textSecondary, fontSize: 12),
                        ),
                      ],
                    ),
                    Row(
                      children: [
                         Column(
                           children: [
                             Container(
                               padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                               decoration: BoxDecoration(color: Colors.orange, borderRadius: BorderRadius.circular(16)),
                               child: const Text('VÀO', style: TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold)),
                             ),
                             const SizedBox(height: 4),
                             Text(inStr, style: const TextStyle(color: Colors.orange, fontWeight: FontWeight.bold, fontSize: 16)),
                           ],
                         ),
                         const SizedBox(width: 24),
                         Column(
                           children: [
                             Container(
                               padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                               decoration: BoxDecoration(color: AppColors.primary, borderRadius: BorderRadius.circular(16)),
                               child: const Text('RA', style: TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold)),
                             ),
                             const SizedBox(height: 4),
                             Text(outStr, style: const TextStyle(color: AppColors.primary, fontWeight: FontWeight.bold, fontSize: 16)),
                           ],
                         ),
                      ],
                    ),
                  ],
                );
              },
            );
          },
        ),
      ],
    );
  }

  String _getGreeting() {
    final hour = _currentTime.hour;
    if (hour < 12) return 'Chào buổi sáng';
    if (hour < 18) return 'Chào buổi chiều';
    return 'Chào buổi tối';
  }

  void _showMethodPicker(
    BuildContext context, {
    required bool isCheckIn,
  }) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => BlocProvider.value(
          value: context.read<AttendanceBloc>(),
          child: CheckInScreen(
            isCheckIn: isCheckIn,
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
                // Avatar with gradient ring
                Container(
                  padding: const EdgeInsets.all(4),
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: LinearGradient(
                      colors: [AppColors.primary, AppColors.primaryLight],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                  ),
                  child: CircleAvatar(
                    radius: 48,
                    backgroundColor: Colors.white,
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
                ),
                const SizedBox(height: 16),
                Text(
                  user?.name ?? '',
                  style: theme.textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.w800,
                    color: AppColors.primaryDark,
                    fontSize: 22,
                  ),
                ),
                const SizedBox(height: 6),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppColors.primary.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(
                    user?.position ?? '',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: AppColors.primary,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                const SizedBox(height: 28),

                // Info cards
                _buildInfoItem(context, Icons.badge_outlined, 'Mã NV', user?.employeeCode ?? '', AppColors.secondary),
                _buildInfoItem(context, Icons.email_outlined, 'Email', user?.email ?? '', AppColors.info),
                _buildInfoItem(context, Icons.phone_outlined, 'SĐT', user?.phone ?? '', AppColors.success),
                _buildInfoItem(context, Icons.business_outlined, 'Chi nhánh', user?.branch?.name ?? '', AppColors.primary),
                _buildInfoItem(context, Icons.apartment_outlined, 'Phòng ban', user?.department ?? '', AppColors.statusHalfDay),
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
    Color iconColor,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: iconColor.withOpacity(0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(icon, size: 22, color: iconColor),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppColors.textSecondary,
                        fontSize: 12,
                      ),
                ),
                const SizedBox(height: 2),
                Text(
                  value.isEmpty ? '---' : value,
                  style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: AppColors.textPrimary,
                      ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
