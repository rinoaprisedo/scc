import '../../../core/network/api_result.dart';
import '../domain/app_settings.dart';
import 'settings_api.dart';

class SettingsRepository {
  final SettingsApi _api;
  SettingsRepository(this._api);

  Future<ApiResult<AppSettings>> fetchPublicSettings() {
    return guardApiCall(() async {
      final json = await _api.getPublicSettings();
      return AppSettings.fromJson(json);
    });
  }
}
