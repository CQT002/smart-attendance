import '../../data/models/login_response_model.dart';
import '../../data/models/user_model.dart';

abstract class AuthRepository {
  Future<LoginResponseModel> login(String email, String password);
  Future<UserModel> getMe();
  Future<void> changePassword(String oldPassword, String newPassword);
  Future<void> logout();
  Future<bool> isLoggedIn();
}
