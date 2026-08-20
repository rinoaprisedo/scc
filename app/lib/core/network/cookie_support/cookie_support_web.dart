import 'package:dio/browser.dart';
import 'package:dio/dio.dart';
import 'package:web/web.dart' as web;

/// Web: the browser owns cookie storage natively (Set-Cookie / Cookie
/// headers are handled by the browser's own fetch/XHR implementation), so no
/// cookie_jar/path_provider is needed here — just opt into sending
/// credentials cross-origin, and read the non-httpOnly csrf_token cookie out
/// of document.cookie to echo it back as a header.
Future<void> configureCookieSupport(Dio dio, String baseUrl) async {
  final adapter = dio.httpClientAdapter;
  if (adapter is BrowserHttpClientAdapter) {
    adapter.withCredentials = true;
  }
  dio.interceptors.add(_WebCsrfInterceptor());
}

class _WebCsrfInterceptor extends Interceptor {
  static const _safeMethods = {'GET', 'HEAD', 'OPTIONS'};

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    if (!_safeMethods.contains(options.method.toUpperCase())) {
      final match = RegExp(r'csrf_token=([^;]+)').firstMatch(web.document.cookie);
      if (match != null) {
        options.headers['X-CSRF-Token'] = Uri.decodeComponent(match.group(1)!);
      }
    }
    handler.next(options);
  }
}
