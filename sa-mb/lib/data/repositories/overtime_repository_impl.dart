import '../../core/constants/api_constants.dart';
import '../../core/network/api_client.dart';
import '../../domain/repositories/overtime_repository.dart';
import '../models/overtime_model.dart';

class OvertimeRepositoryImpl implements OvertimeRepository {
  final ApiClient _apiClient;

  OvertimeRepositoryImpl(this._apiClient);

  @override
  Future<OvertimeCheckInResponse> checkIn() async {
    final response = await _apiClient.post(ApiConstants.overtimeCheckIn);
    final data = response.data['data'] as Map<String, dynamic>;
    return OvertimeCheckInResponse.fromJson(data);
  }

  @override
  Future<OvertimeCheckOutResponse> checkOut({int? overtimeId}) async {
    final body = <String, dynamic>{};
    if (overtimeId != null) body['overtime_id'] = overtimeId;
    final response = await _apiClient.post(ApiConstants.overtimeCheckOut, data: body);
    final data = response.data['data'] as Map<String, dynamic>;
    return OvertimeCheckOutResponse.fromJson(data);
  }

  @override
  Future<OvertimeModel?> getMyToday() async {
    try {
      final response = await _apiClient.get(ApiConstants.overtimeToday);
      final data = response.data['data'] as Map<String, dynamic>?;
      if (data == null) return null;
      return OvertimeModel.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  @override
  Future<List<OvertimeModel>> getMyList({
    String? status,
    int page = 1,
    int limit = 20,
  }) async {
    final params = <String, dynamic>{'page': page, 'limit': limit};
    if (status != null && status.isNotEmpty) params['status'] = status;

    final response = await _apiClient.get(
      ApiConstants.overtime,
      queryParameters: params,
    );
    final list = response.data['data'] as List<dynamic>? ?? [];
    return list
        .map((e) => OvertimeModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  @override
  Future<OvertimeModel> getById(int id) async {
    final response = await _apiClient.get('${ApiConstants.overtime}/$id');
    final data = response.data['data'] as Map<String, dynamic>;
    return OvertimeModel.fromJson(data);
  }

  @override
  Future<List<OvertimeModel>> getAdminList({
    String? status,
    int page = 1,
    int limit = 20,
  }) async {
    final params = <String, dynamic>{'page': page, 'limit': limit};
    if (status != null && status.isNotEmpty) params['status'] = status;

    final response = await _apiClient.get(
      ApiConstants.adminOvertime,
      queryParameters: params,
    );
    final list = response.data['data'] as List<dynamic>? ?? [];
    return list
        .map((e) => OvertimeModel.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  @override
  Future<OvertimeModel> process({
    required int overtimeId,
    required String status,
    String managerNote = '',
  }) async {
    final response = await _apiClient.put(
      '${ApiConstants.adminOvertime}/$overtimeId/process',
      data: {'status': status, 'manager_note': managerNote},
    );
    final data = response.data['data'] as Map<String, dynamic>;
    return OvertimeModel.fromJson(data);
  }

  @override
  Future<int> batchApprove() async {
    final response = await _apiClient.post(ApiConstants.batchApproveOvertime);
    final data = response.data['data'] as Map<String, dynamic>?;
    return data?['approved_count'] as int? ?? 0;
  }
}
