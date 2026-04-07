import '../../data/models/leave_model.dart';

abstract class LeaveRepository {
  /// Employee: tạo yêu cầu nghỉ phép
  Future<LeaveModel> createLeave({
    required String leaveDate,
    required String leaveType,
    required String description,
  });

  /// Employee: lấy danh sách yêu cầu của bản thân
  Future<List<LeaveModel>> getMyLeaves({
    String? status,
    int page = 1,
    int limit = 20,
  });

  /// Employee: xem chi tiết yêu cầu
  Future<LeaveModel> getLeaveById(int id);

  /// Manager: lấy danh sách yêu cầu cần duyệt (chi nhánh mình)
  Future<List<LeaveModel>> getAdminLeaves({
    String? status,
    int page = 1,
    int limit = 20,
  });

  /// Manager: duyệt hoặc từ chối yêu cầu
  Future<LeaveModel> processLeave({
    required int leaveId,
    required String status,
    String managerNote = '',
  });

  /// Manager: duyệt tất cả yêu cầu đang chờ
  Future<int> batchApprove();
}
