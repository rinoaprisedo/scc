import 'package:dio/dio.dart';

import 'api_exception.dart';

/// Uniform success/failure wrapper for repository calls, so presentation
/// code doesn't need to know about Dio/HTTP details.
sealed class ApiResult<T> {
  const ApiResult();

  factory ApiResult.success(T data) = ApiSuccess<T>;
  factory ApiResult.failure(ApiException error) = ApiFailure<T>;
}

class ApiSuccess<T> extends ApiResult<T> {
  final T data;
  const ApiSuccess(this.data);
}

class ApiFailure<T> extends ApiResult<T> {
  final ApiException error;
  const ApiFailure(this.error);
}

/// Runs [action], mapping DioExceptions/unexpected errors into ApiResult.
Future<ApiResult<T>> guardApiCall<T>(Future<T> Function() action) async {
  try {
    return ApiResult.success(await action());
  } on DioException catch (e) {
    final status = e.response?.statusCode;
    final message = e.response?.data is Map
        ? (e.response?.data['message'] as String? ?? 'Request failed')
        : (e.message ?? 'Request failed');
    return ApiResult.failure(ApiException(message, statusCode: status));
  } catch (e) {
    return ApiResult.failure(ApiException(e.toString()));
  }
}
