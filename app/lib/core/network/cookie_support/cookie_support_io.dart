import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';
import 'package:path_provider/path_provider.dart';

/// Mobile/desktop: a real file-backed cookie jar replays the `session` +
/// `csrf_token` cookies (Dio has no cookie jar of its own, unlike a browser).
Future<void> configureCookieSupport(Dio dio, String baseUrl) async {
  final dir = await getApplicationDocumentsDirectory();
  final jar = PersistCookieJar(storage: FileStorage('${dir.path}/.cookies'));

  dio.interceptors.add(CookieManager(jar));
  dio.interceptors.add(_CsrfInterceptor(jar, baseUrl));
}

class _CsrfInterceptor extends Interceptor {
  final PersistCookieJar jar;
  final String baseUrl;

  _CsrfInterceptor(this.jar, this.baseUrl);

  static const _safeMethods = {'GET', 'HEAD', 'OPTIONS'};

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    if (!_safeMethods.contains(options.method.toUpperCase())) {
      final cookies = await jar.loadForRequest(Uri.parse(baseUrl));
      final csrf = cookies.where((c) => c.name == 'csrf_token').firstOrNull;
      if (csrf != null) {
        options.headers['X-CSRF-Token'] = csrf.value;
      }
    }
    handler.next(options);
  }
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull => isEmpty ? null : first;
}
