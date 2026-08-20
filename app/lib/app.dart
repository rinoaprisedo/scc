import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/router/app_router.dart';
import 'core/theme/app_colors.dart';
import 'core/theme/app_theme.dart';
import 'features/settings/presentation/settings_provider.dart';

class App extends ConsumerWidget {
  const App({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsProvider);
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'BaseAdmin',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.build(primary: settings.value?.primaryColor ?? AppColors.defaultPrimary),
      routerConfig: router,
      builder: (context, child) {
        // Show a lightweight splash while branding/settings load, so there's
        // no flash of the fallback color before the real primary applies.
        if (settings.isLoading) {
          return const Scaffold(
            backgroundColor: AppColors.background,
            body: Center(child: CircularProgressIndicator()),
          );
        }
        return child ?? const SizedBox.shrink();
      },
    );
  }
}
