import 'package:flutter/material.dart';

/// Neutral palette for the clean-white base. The accent (primary) color is
/// dynamic — sourced from the backend's `primary_color` appearance setting,
/// see features/settings.
class AppColors {
  AppColors._();

  static const background = Colors.white;
  static const surface = Colors.white;
  static const border = Color(0xFFE5E7EB);
  static const textPrimary = Color(0xFF111827);
  static const textSecondary = Color(0xFF6B7280);
  static const danger = Color(0xFFDC2626);

  // Matches the React admin frontend's default primary, for visual parity
  // when the backend hasn't set primary_color yet.
  static const defaultPrimary = Color(0xFFC2622E);
}
