import '../../data/models/attendance_model.dart';

abstract class AttendanceRepository {
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
  });

  Future<AttendanceModel> checkOut({
    required int attendanceId,
    required double latitude,
    required double longitude,
    required String ssid,
    required String bssid,
    required String deviceId,
    required bool isFakeGps,
    required bool isVpn,
  });

  Future<AttendanceModel?> getTodayAttendance();

  Future<List<AttendanceModel>> getHistory({
    required DateTime from,
    required DateTime to,
    int page = 1,
    int limit = 20,
  });
}
