import 'package:flutter/material.dart';

import '../../../core/widgets/placeholder_screen.dart';

class CalendarScreen extends StatelessWidget {
  const CalendarScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const PlaceholderScreen(
      title: 'Calendar',
      icon: Icons.calendar_today_rounded,
      message: 'Calendar/schedule content goes here.',
    );
  }
}
