import '../../core/network/api_client.dart';
import '../../core/constants/api_constants.dart';
import '../../data/models/api_response_model.dart';
import '../../data/models/login_response_model.dart';
import '../../data/models/user_model.dart';
import '../../domain/repositories/auth_repository.dart';

class AuthRepositoryImpl implements AuthRepository {
  final ApiClient _apiClient;

  AuthRepositoryImpl(this._apiClient);

  @override
  Future<LoginResponseModel> login(String email, String password) async {
    final response = await _apiClient.post(
      ApiConstants.login,
      data: {'email': email, 'password': password},
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => LoginResponseModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Đăng nhập thất bại');
    }

    // Save tokens
    await _apiClient.saveTokens(
      apiResponse.data!.accessToken,
      apiResponse.data!.refreshToken,
    );

    return apiResponse.data!;
  }

  @override
  Future<UserModel> getMe() async {
    final response = await _apiClient.get(ApiConstants.me);

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      (data) => UserModel.fromJson(data as Map<String, dynamic>),
    );

    if (!apiResponse.success || apiResponse.data == null) {
      throw Exception(apiResponse.error?.message ?? 'Không thể lấy thông tin');
    }

    return apiResponse.data!;
  }

  @override
  Future<void> changePassword(String oldPassword, String newPassword) async {
    final response = await _apiClient.put(
      ApiConstants.changePassword,
      data: {
        'old_password': oldPassword,
        'new_password': newPassword,
      },
    );

    final apiResponse = ApiResponse.fromJson(
      response.data as Map<String, dynamic>,
      null,
    );

    if (!apiResponse.success) {
      throw Exception(apiResponse.error?.message ?? 'Đổi mật khẩu thất bại');
    }
  }

  @override
  Future<void> logout() async {
    await _apiClient.clearTokens();
  }

  @override
  Future<bool> isLoggedIn() async {
    final token = await _apiClient.getAccessToken();
    return token != null && token.isNotEmpty;
  }
}
