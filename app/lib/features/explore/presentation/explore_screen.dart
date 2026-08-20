import 'package:flutter/material.dart';

import '../../../core/widgets/placeholder_screen.dart';

class ExploreScreen extends StatelessWidget {
  const ExploreScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const PlaceholderScreen(
      title: 'Explore',
      icon: Icons.explore_rounded,
      message: 'Explore content goes here.',
    );
  }
}
