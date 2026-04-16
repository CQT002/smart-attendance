import '../../data/models/holiday_model.dart';

abstract class HolidayRepository {
  /// Employee: lấy calendar ngày lễ trong khoảng (nếu bỏ trống → năm hiện tại)
  Future<List<HolidayModel>> getCalendar({String? dateFrom, String? dateTo});

  /// Employee: báo cáo công kèm ngày lễ & hệ số lương
  Future<AttendanceSummaryModel> getAttendanceSummary({
    required String dateFrom,
    required String dateTo,
  });

  /// Admin/Manager: list holidays (admin page)
  Future<List<HolidayModel>> getAdminHolidays({int? year, String? type});

  /// Admin only: tạo ngày lễ
  Future<HolidayModel> createHoliday({
    required String name,
    required String date,
    double coefficient = 0,
    String type = 'national',
    bool isCompensated = false,
    String? compensateFor,
    String description = '',
  });

  /// Admin only: cập nhật
  Future<HolidayModel> updateHoliday({
    required int id,
    required String name,
    required String date,
    double coefficient = 0,
    String type = 'national',
    bool isCompensated = false,
    String? compensateFor,
    String description = '',
  });

  /// Admin only: xoá
  Future<void> deleteHoliday(int id);
}
