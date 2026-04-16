import '../../core/constants/api_constants.dart';
import '../../core/network/api_client.dart';
import '../../data/models/api_response_model.dart';
import '../../data/models/holiday_model.dart';
import '../../domain/repositories/holiday_repository.dart';

class HolidayRepositoryImpl implements HolidayRepository {
  final ApiClient _apiClient;

  HolidayRepositoryImpl(this._apiClient);

  @override
  Future<List<HolidayModel>> getCalendar({String? dateFrom, String? dateTo}) async {
    final params = <String, dynamic>{};
    if (dateFrom != null && dateFrom.isNotEmpty) params['date_from'] = dateFrom;
    if (dateTo != null && dateTo.isNotEmpty) params['date_to'] = dateTo;

    final response = await _apiClient.get(
      ApiConstants.holidayCalendar,
      queryParameters: params.isEmpty ? null : params,
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) {
        final list = data as List<dynamic>;
        return list.map((e) => HolidayModel.fromJson(e as Map<String, dynamic>)).toList();
      },
    );

    if (!apiResponse.success || apiResponse.data == null) {
      return [];
    }
    return apiResponse.data!;
  }

  @override
  Future<AttendanceSummaryModel> getAttendanceSummary({
    required String dateFrom,
    required String dateTo,
  }) async {
    final response = await _apiClient.get(
      ApiConstants.attendanceSummary,
      queryParameters: {'date_from': dateFrom, 'date_to': dateTo},
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => AttendanceSummaryModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Không lấy được báo cáo công');
    }
    return apiResponse.data!;
  }

  @override
  Future<List<HolidayModel>> getAdminHolidays({int? year, String? type}) async {
    final params = <String, dynamic>{};
    if (year != null) params['year'] = year;
    if (type != null && type.isNotEmpty) params['type'] = type;

    final response = await _apiClient.get(
      ApiConstants.adminHolidays,
      queryParameters: params.isEmpty ? null : params,
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) {
        final list = data as List<dynamic>;
        return list.map((e) => HolidayModel.fromJson(e as Map<String, dynamic>)).toList();
      },
    );

    if (!apiResponse.success || apiResponse.data == null) {
      return [];
    }
    return apiResponse.data!;
  }

  @override
  Future<HolidayModel> createHoliday({
    required String name,
    required String date,
    double coefficient = 0,
    String type = 'national',
    bool isCompensated = false,
    String? compensateFor,
    String description = '',
  }) async {
    final response = await _apiClient.post(
      ApiConstants.adminHolidays,
      data: {
        'name': name,
        'date': date,
        'coefficient': coefficient,
        'type': type,
        'is_compensated': isCompensated,
        if (compensateFor != null) 'compensate_for': compensateFor,
        'description': description,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => HolidayModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Tạo ngày lễ thất bại');
    }
    return apiResponse.data!;
  }

  @override
  Future<HolidayModel> updateHoliday({
    required int id,
    required String name,
    required String date,
    double coefficient = 0,
    String type = 'national',
    bool isCompensated = false,
    String? compensateFor,
    String description = '',
  }) async {
    final response = await _apiClient.put(
      '${ApiConstants.adminHolidays}/$id',
      data: {
        'name': name,
        'date': date,
        'coefficient': coefficient,
        'type': type,
        'is_compensated': isCompensated,
        if (compensateFor != null) 'compensate_for': compensateFor,
        'description': description,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => HolidayModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Cập nhật ngày lễ thất bại');
    }
    return apiResponse.data!;
  }

  @override
  Future<void> deleteHoliday(int id) async {
    await _apiClient.delete('${ApiConstants.adminHolidays}/$id');
  }
}
