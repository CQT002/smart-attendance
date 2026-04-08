import '../../data/models/overtime_model.dart';

abstract class OvertimeRepository {
  /// Employee: check-in tăng ca
  Future<OvertimeCheckInResponse> checkIn();

  /// Employee: check-out tăng ca
  Future<OvertimeCheckOutResponse> checkOut({int? overtimeId});

  /// Employee: lấy trạng thái OT hôm nay
  Future<OvertimeModel?> getMyToday();

  /// Employee: lấy lịch sử tăng ca
  Future<List<OvertimeModel>> getMyList({
    String? status,
    int page = 1,
    int limit = 20,
  });

  /// Employee: xem chi tiết yêu cầu OT
  Future<OvertimeModel> getById(int id);

  /// Manager: lấy danh sách yêu cầu OT
  Future<List<OvertimeModel>> getAdminList({
    String? status,
    int page = 1,
    int limit = 20,
  });

  /// Manager: duyệt hoặc từ chối yêu cầu OT
  Future<OvertimeModel> process({
    required int overtimeId,
    required String status,
    String managerNote = '',
  });

  /// Manager: duyệt tất cả yêu cầu OT đang chờ
  Future<int> batchApprove();
}
