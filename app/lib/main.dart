import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app.dart';
import 'core/network/dio_client.dart';
import 'core/network/dio_provider.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final dioClient = await DioClient.create();

  runApp(
    ProviderScope(
      overrides: [dioClientProvider.overrideWithValue(dioClient)],
      child: const App(),
    ),
  );
}
