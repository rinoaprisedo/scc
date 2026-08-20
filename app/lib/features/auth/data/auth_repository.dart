import '../../../core/network/api_result.dart';
import '../domain/user.dart';
import 'auth_api.dart';

class AuthRepository {
  final AuthApi _api;
  AuthRepository(this._api);

  Future<ApiResult<User>> login({required String email, required String password}) {
    return guardApiCall(() async => User.fromJson(await _api.login(email: email, password: password)));
  }

  Future<ApiResult<User>> me() {
    return guardApiCall(() async => User.fromJson(await _api.me()));
  }

  Future<ApiResult<void>> logout() {
    return guardApiCall(() => _api.logout());
  }
}
