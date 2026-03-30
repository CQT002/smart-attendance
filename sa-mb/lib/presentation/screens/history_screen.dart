import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../core/theme/app_colors.dart';
import '../../core/utils/date_utils.dart';
import '../blocs/attendance/attendance_bloc.dart';
import '../blocs/attendance/attendance_event.dart';
import '../blocs/attendance/attendance_state.dart';
import '../widgets/attendance_card.dart';

class HistoryScreen extends StatefulWidget {
  const HistoryScreen({super.key});

  @override
  State<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends State<HistoryScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _tabController.addListener(_onTabChanged);
    _loadData();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void _onTabChanged() {
    if (!_tabController.indexIsChanging) return;
    _loadData();
  }

  void _loadData() {
    final now = DateTime.now();
    DateTime from;
    DateTime to = AppDateUtils.endOfDay(now);

    switch (_tabController.index) {
      case 0: // This week
        from = AppDateUtils.startOfWeek(now);
        break;
      case 1: // This month
        from = AppDateUtils.startOfMonth(now);
        break;
      case 2: // Last month
        final lastMonth = DateTime(now.year, now.month - 1, 1);
        from = AppDateUtils.startOfMonth(lastMonth);
        to = AppDateUtils.endOfMonth(lastMonth);
        break;
      default:
        from = AppDateUtils.startOfWeek(now);
    }

    context.read<AttendanceBloc>().add(
          AttendanceLoadHistory(from: from, to: to),
        );
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 16, 20, 0),
            child: Text(
              'Lich su cham cong',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
          ),
          const SizedBox(height: 16),
          TabBar(
            controller: _tabController,
            labelColor: AppColors.primary,
            unselectedLabelColor: AppColors.textSecondary,
            indicatorColor: AppColors.primary,
            tabs: const [
              Tab(text: 'Tuan nay'),
              Tab(text: 'Thang nay'),
              Tab(text: 'Thang truoc'),
            ],
          ),
          Expanded(
            child: BlocBuilder<AttendanceBloc, AttendanceState>(
              builder: (context, state) {
                if (state is AttendanceLoading) {
                  return const Center(child: CircularProgressIndicator());
                }

                if (state is AttendanceFailure) {
                  return Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.error_outline,
                            size: 48, color: AppColors.error),
                        const SizedBox(height: 16),
                        Text(state.message),
                        const SizedBox(height: 16),
                        TextButton(
                          onPressed: _loadData,
                          child: const Text('Thu lai'),
                        ),
                      ],
                    ),
                  );
                }

                if (state is AttendanceHistoryLoaded) {
                  if (state.records.isEmpty) {
                    return Center(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.event_busy,
                              size: 48,
                              color: AppColors.textSecondary.withOpacity(0.5)),
                          const SizedBox(height: 16),
                          Text(
                            'Khong co du lieu cham cong',
                            style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                                  color: AppColors.textSecondary,
                                ),
                          ),
                        ],
                      ),
                    );
                  }

                  return ListView.builder(
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    itemCount: state.records.length,
                    itemBuilder: (context, index) {
                      return AttendanceCard(attendance: state.records[index]);
                    },
                  );
                }

                return const SizedBox.shrink();
              },
            ),
          ),
        ],
      ),
    );
  }
}
