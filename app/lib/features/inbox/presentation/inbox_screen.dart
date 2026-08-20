import 'package:flutter/material.dart';

import '../../../core/widgets/placeholder_screen.dart';

class InboxScreen extends StatelessWidget {
  const InboxScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const PlaceholderScreen(
      title: 'Inbox',
      icon: Icons.mail_rounded,
      message: 'Messages/notifications content goes here.',
    );
  }
}
