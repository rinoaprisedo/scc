import 'package:dio/dio.dart';

class SettingsApi {
  final Dio _dio;
  SettingsApi(this._dio);

  /// GET /settings/public — unauthenticated, whitelisted branding fields.
  /// Also doubles as the request that seeds the csrf_token cookie before
  /// the first mutating request (e.g. login).
  Future<Map<String, dynamic>> getPublicSettings() async {
    final res = await _dio.get('/settings/public');
    return Map<String, dynamic>.from(res.data['data'] as Map);
  }
}
