import '../../core/network/api_client.dart';
import '../../core/constants/api_constants.dart';
import '../../data/models/api_response_model.dart';
import '../../data/models/leave_model.dart';
import '../../domain/repositories/leave_repository.dart';

class LeaveRepositoryImpl implements LeaveRepository {
  final ApiClient _apiClient;

  LeaveRepositoryImpl(this._apiClient);

  @override
  Future<LeaveModel> createLeave({
    required String leaveDate,
    required String leaveType,
    required String description,
  }) async {
    final response = await _apiClient.post(
      ApiConstants.leaves,
      data: {
        'leave_date': leaveDate,
        'leave_type': leaveType,
        'description': description,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => LeaveModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Tạo yêu cầu nghỉ phép thất bại');
    }

    return apiResponse.data!;
  }

  @override
  Future<List<LeaveModel>> getMyLeaves({
    String? status,
    int page = 1,
    int limit = 20,
  }) async {
    final params = <String, dynamic>{
      'page': page,
      'limit': limit,
    };
    if (status != null && status.isNotEmpty) {
      params['status'] = status;
    }

    final response = await _apiClient.get(
      ApiConstants.leaves,
      queryParameters: params,
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) {
        final list = data as List<dynamic>;
        return list
            .map((item) => LeaveModel.fromJson(item as Map<String, dynamic>))
            .toList();
      },
    );

    if (!apiResponse.success || apiResponse.data == null) {
      return [];
    }

    return apiResponse.data!;
  }

  @override
  Future<LeaveModel> getLeaveById(int id) async {
    final response = await _apiClient.get('${ApiConstants.leaves}/$id');

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => LeaveModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Không tìm thấy yêu cầu');
    }

    return apiResponse.data!;
  }

  @override
  Future<List<LeaveModel>> getAdminLeaves({
    String? status,
    int page = 1,
    int limit = 20,
  }) async {
    final params = <String, dynamic>{
      'page': page,
      'limit': limit,
    };
    if (status != null && status.isNotEmpty) {
      params['status'] = status;
    }

    final response = await _apiClient.get(
      ApiConstants.adminLeaves,
      queryParameters: params,
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) {
        final list = data as List<dynamic>;
        return list
            .map((item) => LeaveModel.fromJson(item as Map<String, dynamic>))
            .toList();
      },
    );

    if (!apiResponse.success || apiResponse.data == null) {
      return [];
    }

    return apiResponse.data!;
  }

  @override
  Future<LeaveModel> processLeave({
    required int leaveId,
    required String status,
    String managerNote = '',
  }) async {
    final response = await _apiClient.put(
      '${ApiConstants.adminLeaves}/$leaveId/process',
      data: {
        'status': status,
        'manager_note': managerNote,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => LeaveModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Xử lý yêu cầu thất bại');
    }

    return apiResponse.data!;
  }

  @override
  Future<int> batchApprove() async {
    final response = await _apiClient.post(
      ApiConstants.batchApproveLeaves,
    );

    final data = response.data as Map<String, dynamic>;
    final result = data['data'] as Map<String, dynamic>?;
    return result?['approved_count'] as int? ?? 0;
  }
}
