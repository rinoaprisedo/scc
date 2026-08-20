import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/auth_provider.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/calendar/presentation/calendar_screen.dart';
import '../../features/explore/presentation/explore_screen.dart';
import '../../features/home/presentation/home_screen.dart';
import '../../features/inbox/presentation/inbox_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../shell/main_shell.dart';
import 'routes.dart';

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: Routes.home,
    refreshListenable: GoRouterRefreshStream(ref),
    redirect: (context, state) {
      final authState = ref.read(authProvider);
      final isLoading = authState.isLoading;
      final isLoggedIn = authState.value != null;
      final isLoggingIn = state.matchedLocation == Routes.login;

      if (isLoading) return null; // stay put while session check is in flight
      if (!isLoggedIn && !isLoggingIn) return Routes.login;
      if (isLoggedIn && isLoggingIn) return Routes.home;
      return null;
    },
    routes: [
      GoRoute(path: Routes.login, builder: (context, state) => const LoginScreen()),
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) => MainShell(navigationShell: navigationShell),
        branches: [
          StatefulShellBranch(routes: [
            GoRoute(path: Routes.home, builder: (context, state) => const HomeScreen()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: Routes.explore, builder: (context, state) => const ExploreScreen()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: Routes.calendar, builder: (context, state) => const CalendarScreen()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: Routes.inbox, builder: (context, state) => const InboxScreen()),
          ]),
        ],
      ),
      // Reached via the "More" sheet (shell/more_menu_sheet.dart), not a
      // bottom-nav tab — pushed on top of the shell rather than replacing it.
      GoRoute(path: Routes.profile, builder: (context, state) => const ProfileScreen()),
    ],
  );
});

/// Lets go_router's `redirect` re-run whenever authProvider's AsyncValue
/// changes (login/logout/session-check completing).
class GoRouterRefreshStream extends ChangeNotifier {
  GoRouterRefreshStream(Ref ref) {
    ref.listen(authProvider, (previous, next) => notifyListeners());
  }
}
