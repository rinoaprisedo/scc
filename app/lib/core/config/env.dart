enum Env { dev, prod }

/// API base URL differs per environment. Selected at build/run time via
/// `--dart-define=ENV=prod` (defaults to dev when omitted).
class AppConfig {
  final Env env;
  final String baseUrl;

  const AppConfig._(this.env, this.baseUrl);

  // 10.0.2.2 is the Android emulator's alias for the host machine's
  // localhost. iOS simulators can reach the host directly via localhost.
  static const dev = AppConfig._(Env.dev, 'http://10.0.2.2:8500/api/v1');

  // Placeholder — replace with the real production API host before release.
  static const prod = AppConfig._(Env.prod, 'https://api.yourapp.com/api/v1');

  static const _envName = String.fromEnvironment('ENV', defaultValue: 'dev');

  // Escape hatch for targets where the emulator-only 10.0.2.2 alias doesn't
  // resolve (web, desktop, iOS simulator): --dart-define=API_BASE_URL=...
  static const _baseUrlOverride = String.fromEnvironment('API_BASE_URL');

  static AppConfig get current {
    final base = _envName == 'prod' ? prod : dev;
    if (_baseUrlOverride.isEmpty) return base;
    return AppConfig._(base.env, _baseUrlOverride);
  }
}
