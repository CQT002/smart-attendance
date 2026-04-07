import '../../data/models/correction_model.dart';

abstract class CorrectionRepository {
  /// Employee: tạo yêu cầu chấm công bù
  Future<CorrectionModel> createCorrection({
    required int attendanceLogId,
    required String description,
  });

  /// Employee: lấy danh sách yêu cầu của bản thân
  Future<List<CorrectionModel>> getMyCorrections({
    String? status,
    int page = 1,
    int limit = 20,
  });

  /// Employee: xem chi tiết yêu cầu
  Future<CorrectionModel> getCorrectionById(int id);

  /// Manager: lấy danh sách yêu cầu cần duyệt (chi nhánh mình)
  Future<List<CorrectionModel>> getAdminCorrections({
    String? status,
    int page = 1,
    int limit = 20,
  });

  /// Manager: duyệt hoặc từ chối yêu cầu
  Future<CorrectionModel> processCorrection({
    required int correctionId,
    required String status, // 'approved' or 'rejected'
    String managerNote = '',
  });
}
