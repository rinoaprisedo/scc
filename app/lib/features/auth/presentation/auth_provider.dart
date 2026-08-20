import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_result.dart';
import '../../../core/network/dio_provider.dart';
import '../data/auth_api.dart';
import '../data/auth_repository.dart';
import '../domain/user.dart';

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final dio = ref.watch(dioClientProvider).dio;
  return AuthRepository(AuthApi(dio));
});

/// Drives router redirects: null (loading) -> splash, null (data, no error)
/// with hasValue false -> unauthenticated, User -> authenticated.
class AuthNotifier extends AsyncNotifier<User?> {
  @override
  Future<User?> build() async {
    // App boot: if a valid session cookie survived from a previous run,
    // GET /auth/me succeeds and we land straight on Home.
    final result = await ref.read(authRepositoryProvider).me();
    return switch (result) {
      ApiSuccess(:final data) => data,
      ApiFailure() => null,
    };
  }

  Future<String?> login({required String email, required String password}) async {
    state = const AsyncLoading();
    final result = await ref.read(authRepositoryProvider).login(email: email, password: password);
    return switch (result) {
      ApiSuccess(:final data) => () {
          state = AsyncData(data);
          return null;
        }(),
      ApiFailure(:final error) => () {
          state = const AsyncData(null);
          return error.message;
        }(),
    };
  }

  Future<void> logout() async {
    await ref.read(authRepositoryProvider).logout();
    state = const AsyncData(null);
  }
}

final authProvider = AsyncNotifierProvider<AuthNotifier, User?>(AuthNotifier.new);
