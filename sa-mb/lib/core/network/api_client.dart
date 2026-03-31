import 'dart:io' show Platform;

import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../constants/api_constants.dart';
import '../constants/app_constants.dart';

/// Token storage abstraction to handle platform differences.
/// macOS sandbox blocks Keychain without signing, so we fall back
/// to SharedPreferences on desktop platforms.
class _TokenStorage {
  final FlutterSecureStorage? _secure;
  SharedPreferences? _prefs;
  final bool _useFallback;

  _TokenStorage()
      : _useFallback = Platform.isMacOS || Platform.isWindows || Platform.isLinux,
        _secure = (Platform.isMacOS || Platform.isWindows || Platform.isLinux)
            ? null
            : const FlutterSecureStorage();

  Future<void> _ensurePrefs() async {
    _prefs ??= await SharedPreferences.getInstance();
  }

  Future<String?> read(String key) async {
    if (_useFallback) {
      await _ensurePrefs();
      return _prefs!.getString(key);
    }
    return _secure!.read(key: key);
  }

  Future<void> write(String key, String value) async {
    if (_useFallback) {
      await _ensurePrefs();
      await _prefs!.setString(key, value);
    } else {
      await _secure!.write(key: key, value: value);
    }
  }

  Future<void> delete(String key) async {
    if (_useFallback) {
      await _ensurePrefs();
      await _prefs!.remove(key);
    } else {
      await _secure!.delete(key: key);
    }
  }
}

class ApiClient {
  late final Dio _dio;
  final _TokenStorage _storage = _TokenStorage();

  ApiClient() {
    _dio = Dio(
      BaseOptions(
        baseUrl: ApiConstants.baseUrl,
        connectTimeout: const Duration(seconds: 15),
        receiveTimeout: const Duration(seconds: 15),
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
      ),
    );

    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await _storage.read(AppConstants.tokenKey);
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
        onError: (error, handler) async {
          if (error.response?.statusCode == 401) {
            await _storage.delete(AppConstants.tokenKey);
            await _storage.delete(AppConstants.refreshTokenKey);
          }
          handler.next(error);
        },
      ),
    );
  }

  Future<Response> get(String path, {Map<String, dynamic>? queryParameters}) {
    return _dio.get(path, queryParameters: queryParameters);
  }

  Future<Response> post(String path, {dynamic data}) {
    return _dio.post(path, data: data);
  }

  Future<Response> put(String path, {dynamic data}) {
    return _dio.put(path, data: data);
  }

  Future<Response> delete(String path) {
    return _dio.delete(path);
  }

  Future<void> saveTokens(String accessToken, String refreshToken) async {
    await _storage.write(AppConstants.tokenKey, accessToken);
    await _storage.write(AppConstants.refreshTokenKey, refreshToken);
  }

  Future<void> clearTokens() async {
    await _storage.delete(AppConstants.tokenKey);
    await _storage.delete(AppConstants.refreshTokenKey);
  }

  Future<String?> getAccessToken() {
    return _storage.read(AppConstants.tokenKey);
  }
}
