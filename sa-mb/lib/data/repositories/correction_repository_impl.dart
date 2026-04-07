import '../../core/network/api_client.dart';
import '../../core/constants/api_constants.dart';
import '../../data/models/api_response_model.dart';
import '../../data/models/correction_model.dart';
import '../../domain/repositories/correction_repository.dart';

class CorrectionRepositoryImpl implements CorrectionRepository {
  final ApiClient _apiClient;

  CorrectionRepositoryImpl(this._apiClient);

  @override
  Future<CorrectionModel> createCorrection({
    required int attendanceLogId,
    required String description,
  }) async {
    final response = await _apiClient.post(
      ApiConstants.corrections,
      data: {
        'attendance_log_id': attendanceLogId,
        'description': description,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => CorrectionModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Tạo yêu cầu bù công thất bại');
    }

    return apiResponse.data!;
  }

  @override
  Future<List<CorrectionModel>> getMyCorrections({
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
      ApiConstants.corrections,
      queryParameters: params,
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) {
        final list = data as List<dynamic>;
        return list
            .map((item) => CorrectionModel.fromJson(item as Map<String, dynamic>))
            .toList();
      },
    );

    if (!apiResponse.success || apiResponse.data == null) {
      return [];
    }

    return apiResponse.data!;
  }

  @override
  Future<CorrectionModel> getCorrectionById(int id) async {
    final response = await _apiClient.get('${ApiConstants.corrections}/$id');

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => CorrectionModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Không tìm thấy yêu cầu');
    }

    return apiResponse.data!;
  }

  @override
  Future<List<CorrectionModel>> getAdminCorrections({
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
      ApiConstants.adminCorrections,
      queryParameters: params,
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) {
        final list = data as List<dynamic>;
        return list
            .map((item) => CorrectionModel.fromJson(item as Map<String, dynamic>))
            .toList();
      },
    );

    if (!apiResponse.success || apiResponse.data == null) {
      return [];
    }

    return apiResponse.data!;
  }

  @override
  Future<CorrectionModel> processCorrection({
    required int correctionId,
    required String status,
    String managerNote = '',
  }) async {
    final response = await _apiClient.put(
      '${ApiConstants.adminCorrections}/$correctionId/process',
      data: {
        'status': status,
        'manager_note': managerNote,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => CorrectionModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Xử lý yêu cầu thất bại');
    }

    return apiResponse.data!;
  }
}
