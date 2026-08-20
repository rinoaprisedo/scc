import 'package:flutter/material.dart';

import '../theme/app_colors.dart';
import 'floating_title_chip.dart';

/// Shared scaffold for feature stubs — replace with real content per module.
class PlaceholderScreen extends StatelessWidget {
  final String title;
  final IconData icon;
  final String message;

  const PlaceholderScreen({super.key, required this.title, required this.icon, required this.message});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(icon, size: 40, color: AppColors.textSecondary),
                const SizedBox(height: 12),
                Text(message, style: const TextStyle(color: AppColors.textSecondary)),
              ],
            ),
          ),
          FloatingTitleChip(title: title, icon: icon),
        ],
      ),
    );
  }
}
