import 'package:dio/dio.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../domain/repositories/auth_repository.dart';
import 'auth_event.dart';
import 'auth_state.dart';

class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final AuthRepository _authRepository;

  AuthBloc(this._authRepository) : super(AuthInitial()) {
    on<AuthCheckRequested>(_onCheckRequested);
    on<AuthLoginRequested>(_onLoginRequested);
    on<AuthLogoutRequested>(_onLogoutRequested);
    on<AuthRefreshRequested>(_onRefreshRequested);
  }

  Future<void> _onCheckRequested(
    AuthCheckRequested event,
    Emitter<AuthState> emit,
  ) async {
    // Luôn yêu cầu đăng nhập mỗi lần khởi động app (bảo mật cho app chấm công)
    await _authRepository.logout();
    emit(AuthUnauthenticated());
  }

  Future<void> _onLoginRequested(
    AuthLoginRequested event,
    Emitter<AuthState> emit,
  ) async {
    emit(AuthLoading());
    try {
      final loginResponse = await _authRepository.login(
        event.email,
        event.password,
      );
      emit(AuthAuthenticated(loginResponse.user));
    } catch (e) {
      emit(AuthFailure(_extractMessage(e)));
    }
  }

  Future<void> _onLogoutRequested(
    AuthLogoutRequested event,
    Emitter<AuthState> emit,
  ) async {
    await _authRepository.logout();
    emit(AuthUnauthenticated());
  }

  Future<void> _onRefreshRequested(
    AuthRefreshRequested event,
    Emitter<AuthState> emit,
  ) async {
    try {
      final user = await _authRepository.getMe();
      emit(AuthAuthenticated(user));
    } catch (_) {
      await _authRepository.logout();
      emit(AuthUnauthenticated());
    }
  }

  String _extractMessage(Object e) {
    if (e is DioException) {
      final data = e.response?.data;
      if (data is Map<String, dynamic>) {
        final error = data['error'];
        if (error is Map<String, dynamic> && error['message'] != null) {
          return error['message'].toString();
        }
      }
      switch (e.type) {
        case DioExceptionType.connectionTimeout:
        case DioExceptionType.sendTimeout:
        case DioExceptionType.receiveTimeout:
          return 'Kết nối tới máy chủ bị timeout. Vui lòng thử lại.';
        case DioExceptionType.connectionError:
          return 'Không thể kết nối tới máy chủ. Vui lòng kiểm tra mạng.';
        default:
          return 'Đăng nhập thất bại. Vui lòng thử lại.';
      }
    }
    return e.toString().replaceFirst('Exception: ', '');
  }
}
