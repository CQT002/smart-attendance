import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'core/network/api_client.dart';
import 'core/theme/app_theme.dart';
import 'data/repositories/auth_repository_impl.dart';
import 'data/repositories/attendance_repository_impl.dart';
import 'data/repositories/correction_repository_impl.dart';
import 'data/repositories/leave_repository_impl.dart';
import 'data/repositories/overtime_repository_impl.dart';
import 'data/services/device_service.dart';
import 'data/services/location_service.dart';
import 'data/services/security_service.dart';
import 'data/services/wifi_service.dart';
import 'domain/repositories/auth_repository.dart';
import 'domain/repositories/attendance_repository.dart';
import 'domain/repositories/correction_repository.dart';
import 'domain/repositories/leave_repository.dart';
import 'domain/repositories/overtime_repository.dart';
import 'presentation/blocs/auth/auth_bloc.dart';
import 'presentation/blocs/auth/auth_event.dart';
import 'presentation/blocs/auth/auth_state.dart';
import 'presentation/blocs/attendance/attendance_bloc.dart';
import 'presentation/blocs/correction/correction_bloc.dart';
import 'presentation/blocs/leave/leave_bloc.dart';
import 'presentation/blocs/overtime/overtime_bloc.dart';
import 'presentation/screens/login_screen.dart';
import 'presentation/screens/home_screen.dart';

class SmartAttendanceApp extends StatefulWidget {
  const SmartAttendanceApp({super.key});

  @override
  State<SmartAttendanceApp> createState() => _SmartAttendanceAppState();
}

class _SmartAttendanceAppState extends State<SmartAttendanceApp> {
  late final ApiClient _apiClient;
  late final AuthRepository _authRepository;
  late final AttendanceRepository _attendanceRepository;
  late final CorrectionRepository _correctionRepository;
  late final LeaveRepository _leaveRepository;
  late final OvertimeRepository _overtimeRepository;
  late final LocationService _locationService;
  late final WifiService _wifiService;
  late final DeviceService _deviceService;
  late final SecurityService _securityService;

  @override
  void initState() {
    super.initState();
    _apiClient = ApiClient();
    _authRepository = AuthRepositoryImpl(_apiClient);
    _attendanceRepository = AttendanceRepositoryImpl(_apiClient);
    _correctionRepository = CorrectionRepositoryImpl(_apiClient);
    _leaveRepository = LeaveRepositoryImpl(_apiClient);
    _overtimeRepository = OvertimeRepositoryImpl(_apiClient);
    _locationService = LocationService();
    _wifiService = WifiService();
    _deviceService = DeviceService();
    _securityService = SecurityService(_locationService);
  }

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => AuthBloc(_authRepository)..add(AuthCheckRequested()),
      child: MaterialApp(
        title: 'Smart Attendance',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme(),
        themeMode: ThemeMode.light,
        home: BlocBuilder<AuthBloc, AuthState>(
          builder: (context, state) {
            if (state is AuthAuthenticated) {
              return MultiBlocProvider(
                providers: [
                  RepositoryProvider.value(value: _attendanceRepository),
                  RepositoryProvider.value(value: _correctionRepository),
                  RepositoryProvider.value(value: _leaveRepository),
                  RepositoryProvider.value(value: _overtimeRepository),
                  BlocProvider(
                    create: (_) => AttendanceBloc(
                      attendanceRepository: _attendanceRepository,
                      locationService: _locationService,
                      wifiService: _wifiService,
                      deviceService: _deviceService,
                      securityService: _securityService,
                      user: state.user,
                    ),
                  ),
                  BlocProvider(
                    create: (_) => CorrectionBloc(
                      correctionRepository: _correctionRepository,
                    ),
                  ),
                  BlocProvider(
                    create: (_) => LeaveBloc(
                      leaveRepository: _leaveRepository,
                    ),
                  ),
                  BlocProvider(
                    create: (_) => OvertimeBloc(
                      overtimeRepository: _overtimeRepository,
                    ),
                  ),
                ],
                child: const HomeScreen(),
              );
            }

            if (state is AuthInitial) {
              return const Scaffold(
                body: Center(
                  child: CircularProgressIndicator(),
                ),
              );
            }

            // AuthUnauthenticated, AuthLoading, AuthFailure → show login
            return const LoginScreen();
          },
        ),
      ),
    );
  }
}
