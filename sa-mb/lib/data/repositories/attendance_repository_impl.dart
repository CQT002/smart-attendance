import '../../core/network/api_client.dart';
import '../../core/constants/api_constants.dart';
import '../../data/models/api_response_model.dart';
import '../../data/models/attendance_model.dart';
import '../../domain/repositories/attendance_repository.dart';

class AttendanceRepositoryImpl implements AttendanceRepository {
  final ApiClient _apiClient;

  AttendanceRepositoryImpl(this._apiClient);

  @override
  Future<AttendanceModel> checkIn({
    required int branchId,
    required double latitude,
    required double longitude,
    required String ssid,
    required String bssid,
    required String deviceId,
    required String deviceModel,
    required String appVersion,
    required bool isFakeGps,
    required bool isVpn,
  }) async {
    final response = await _apiClient.post(
      ApiConstants.checkIn,
      data: {
        'branch_id': branchId,
        'latitude': latitude,
        'longitude': longitude,
        'ssid': ssid,
        'bssid': bssid,
        'device_id': deviceId,
        'device_model': deviceModel,
        'app_version': appVersion,
        'is_fake_gps': isFakeGps,
        'is_vpn': isVpn,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => AttendanceModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Check-in thất bại');
    }

    return apiResponse.data!;
  }

  @override
  Future<AttendanceModel> checkOut({
    required double latitude,
    required double longitude,
    required String ssid,
    required String bssid,
    required String deviceId,
    required bool isFakeGps,
    required bool isVpn,
  }) async {
    final response = await _apiClient.post(
      ApiConstants.checkOut,
      data: {
        'latitude': latitude,
        'longitude': longitude,
        'ssid': ssid,
        'bssid': bssid,
        'device_id': deviceId,
        'is_fake_gps': isFakeGps,
        'is_vpn': isVpn,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => AttendanceModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Check-out thất bại');
    }

    return apiResponse.data!;
  }

  @override
  Future<AttendanceModel?> getTodayAttendance() async {
    try {
      final response = await _apiClient.get(ApiConstants.todayAttendance);

      final apiResponse = ApiResponse.fromJson(
        response.data as Map<String, dynamic>,
        (data) => AttendanceModel.fromJson(data as Map<String, dynamic>),
      );

      if (!apiResponse.success) return null;
      return apiResponse.data;
    } catch (_) {
      return null;
    }
  }

  @override
  Future<List<AttendanceModel>> getHistory({
    required DateTime from,
    required DateTime to,
    int page = 1,
    int limit = 20,
  }) async {
    final response = await _apiClient.get(
      ApiConstants.attendanceHistory,
      queryParameters: {
        'date_from': from.toIso8601String().split('T').first,
        'date_to': to.toIso8601String().split('T').first,
        'page': page,
        'limit': limit,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) {
        final list = data as List<dynamic>;
        return list
            .map((item) => AttendanceModel.fromJson(item as Map<String, dynamic>))
            .toList();
      },
    );

    if (!apiResponse.success || apiResponse.data == null) {
      return [];
    }

    return apiResponse.data!;
  }
}
