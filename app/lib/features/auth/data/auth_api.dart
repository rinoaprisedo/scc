import 'package:dio/dio.dart';

class AuthApi {
  final Dio _dio;
  AuthApi(this._dio);

  /// POST /auth/login — session comes back as a Set-Cookie header (handled
  /// transparently by the cookie jar), not in this response body.
  Future<Map<String, dynamic>> login({required String email, required String password}) async {
    final res = await _dio.post('/auth/login', data: {'email': email, 'password': password});
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<Map<String, dynamic>> me() async {
    final res = await _dio.get('/auth/me');
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }

  Future<void> logout() async {
    await _dio.post('/auth/logout');
  }
}
