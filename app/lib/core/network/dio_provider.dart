import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'dio_client.dart';

/// Overridden in main() once the async DioClient.create() completes, so
/// every other provider can depend on it synchronously.
final dioClientProvider = Provider<DioClient>((ref) {
  throw UnimplementedError('dioClientProvider must be overridden in main()');
});
