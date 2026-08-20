import 'package:dio/dio.dart';

import '../config/env.dart';
import 'cookie_support/cookie_support.dart';

/// The backend uses httpOnly cookie sessions + a CSRF double-submit cookie
/// (see backend/middleware/{session_auth,csrf}.go) — not bearer tokens. This
/// client behaves like a browser: the session cookie is replayed
/// automatically on every request, and the csrf_token cookie is echoed back
/// as the `X-CSRF-Token` header on mutating requests. The exact mechanism
/// (a real cookie jar vs. the browser's native cookie handling) differs by
/// platform — see cookie_support/.
class DioClient {
  final Dio dio;

  DioClient._(this.dio);

  static Future<DioClient> create() async {
    final dio = Dio(BaseOptions(
      baseUrl: AppConfig.current.baseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Accept': 'application/json'},
    ));

    await configureCookieSupport(dio, AppConfig.current.baseUrl);

    return DioClient._(dio);
  }
}
