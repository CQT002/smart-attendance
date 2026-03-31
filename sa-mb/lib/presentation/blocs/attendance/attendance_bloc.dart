import 'package:dio/dio.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../data/models/user_model.dart';
import '../../../data/services/device_service.dart';
import '../../../data/services/location_service.dart';
import '../../../data/services/security_service.dart';
import '../../../data/services/wifi_service.dart';
import '../../../domain/repositories/attendance_repository.dart';
import 'attendance_event.dart';
import 'attendance_state.dart';

class AttendanceBloc extends Bloc<AttendanceEvent, AttendanceState> {
  final AttendanceRepository _attendanceRepository;
  final LocationService _locationService;
  final WifiService _wifiService;
  final DeviceService _deviceService;
  final SecurityService _securityService;
  final UserModel _user;

  AttendanceBloc({
    required AttendanceRepository attendanceRepository,
    required LocationService locationService,
    required WifiService wifiService,
    required DeviceService deviceService,
    required SecurityService securityService,
    required UserModel user,
  })  : _attendanceRepository = attendanceRepository,
        _locationService = locationService,
        _wifiService = wifiService,
        _deviceService = deviceService,
        _securityService = securityService,
        _user = user,
        super(AttendanceInitial()) {
    on<AttendanceLoadToday>(_onLoadToday);
    on<AttendanceCheckInRequested>(_onCheckIn);
    on<AttendanceCheckOutRequested>(_onCheckOut);
    on<AttendanceLoadHistory>(_onLoadHistory);
  }

  Future<void> _onLoadToday(
    AttendanceLoadToday event,
    Emitter<AttendanceState> emit,
  ) async {
    emit(AttendanceLoading());
    try {
      final today = await _attendanceRepository.getTodayAttendance();
      emit(AttendanceTodayLoaded(today));
    } catch (e) {
      emit(AttendanceFailure(_extractMessage(e)));
    }
  }

  Future<void> _onCheckIn(
    AttendanceCheckInRequested event,
    Emitter<AttendanceState> emit,
  ) async {
    emit(AttendanceLoading());
    try {
      // 1. Security check
      final securityResult = await _securityService.performSecurityCheck();

      // 2. Get device info
      final deviceInfo = await _deviceService.getDeviceInfo();

      // 3. Get location data based on method
      double latitude = 0;
      double longitude = 0;
      String ssid = '';
      String bssid = '';

      if (event.method == 'gps') {
        final position = await _locationService.getCurrentPosition();
        if (position == null) {
          emit(const AttendanceFailure('Không thể lấy vị trí GPS. Vui lòng bật định vị.'));
          return;
        }
        latitude = position.latitude;
        longitude = position.longitude;
      } else {
        // WiFi method
        final wifiInfo = await _wifiService.getCurrentWifiInfo();
        if (wifiInfo == null) {
          emit(const AttendanceFailure('Không thể lấy thông tin WiFi. Vui lòng kết nối WiFi.'));
          return;
        }
        ssid = wifiInfo.ssid;
        bssid = wifiInfo.bssid;

        // Also get GPS for WiFi check-in
        final position = await _locationService.getCurrentPosition();
        if (position != null) {
          latitude = position.latitude;
          longitude = position.longitude;
        }
      }

      // 4. Call API
      final attendance = await _attendanceRepository.checkIn(
        branchId: _user.branchId ?? 0,
        latitude: latitude,
        longitude: longitude,
        ssid: ssid,
        bssid: bssid,
        deviceId: deviceInfo.deviceId,
        deviceModel: deviceInfo.deviceModel,
        appVersion: deviceInfo.appVersion,
        isFakeGps: securityResult.isFakeGps,
        isVpn: securityResult.isVpn,
      );

      emit(AttendanceCheckInSuccess(attendance));
    } catch (e) {
      emit(AttendanceFailure(_extractMessage(e)));
    }
  }

  Future<void> _onCheckOut(
    AttendanceCheckOutRequested event,
    Emitter<AttendanceState> emit,
  ) async {
    emit(AttendanceLoading());
    try {
      // 1. Security check
      final securityResult = await _securityService.performSecurityCheck();

      // 2. Get device info
      final deviceInfo = await _deviceService.getDeviceInfo();

      // 3. Get location data
      double latitude = 0;
      double longitude = 0;
      String ssid = '';
      String bssid = '';

      if (event.method == 'gps') {
        final position = await _locationService.getCurrentPosition();
        if (position == null) {
          emit(const AttendanceFailure('Không thể lấy vị trí GPS.'));
          return;
        }
        latitude = position.latitude;
        longitude = position.longitude;
      } else {
        final wifiInfo = await _wifiService.getCurrentWifiInfo();
        if (wifiInfo == null) {
          emit(const AttendanceFailure('Không thể lấy thông tin WiFi.'));
          return;
        }
        ssid = wifiInfo.ssid;
        bssid = wifiInfo.bssid;

        final position = await _locationService.getCurrentPosition();
        if (position != null) {
          latitude = position.latitude;
          longitude = position.longitude;
        }
      }

      // 4. Call API
      final attendance = await _attendanceRepository.checkOut(
        attendanceId: event.attendanceId,
        latitude: latitude,
        longitude: longitude,
        ssid: ssid,
        bssid: bssid,
        deviceId: deviceInfo.deviceId,
        isFakeGps: securityResult.isFakeGps,
        isVpn: securityResult.isVpn,
      );

      emit(AttendanceCheckOutSuccess(attendance));
    } catch (e) {
      emit(AttendanceFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLoadHistory(
    AttendanceLoadHistory event,
    Emitter<AttendanceState> emit,
  ) async {
    emit(AttendanceLoading());
    try {
      final records = await _attendanceRepository.getHistory(
        from: event.from,
        to: event.to,
        page: event.page,
      );
      emit(AttendanceHistoryLoaded(
        records: records,
        hasMore: records.length >= 20,
      ));
    } catch (e) {
      emit(AttendanceFailure(_extractMessage(e)));
    }
  }

  String _extractMessage(Object e) {
    if (e is DioException) {
      final data = e.response?.data;
      if (data is Map<String, dynamic>) {
        final error = data['error'];
        if (error is Map<String, dynamic> && error['message'] != null) {
          return error['message'].toString();
        }
      }
      switch (e.type) {
        case DioExceptionType.connectionTimeout:
        case DioExceptionType.sendTimeout:
        case DioExceptionType.receiveTimeout:
          return 'Kết nối tới máy chủ bị timeout. Vui lòng thử lại.';
        case DioExceptionType.connectionError:
          return 'Không thể kết nối tới máy chủ. Vui lòng kiểm tra mạng.';
        case DioExceptionType.badResponse:
          final code = e.response?.statusCode;
          if (code == 404) return 'Không tìm thấy dữ liệu.';
          if (code == 401) return 'Phiên đăng nhập đã hết hạn.';
          if (code == 403) return 'Bạn không có quyền thực hiện thao tác này.';
          return 'Lỗi máy chủ ($code). Vui lòng thử lại.';
        default:
          return 'Đã xảy ra lỗi. Vui lòng thử lại.';
      }
    }
    return e.toString().replaceFirst('Exception: ', '');
  }
}
