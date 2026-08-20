import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';

class AppSettings {
  final String appName;
  final String? logoUrl;
  final String? faviconUrl;
  final Color primaryColor;

  const AppSettings({
    required this.appName,
    this.logoUrl,
    this.faviconUrl,
    required this.primaryColor,
  });

  factory AppSettings.fromJson(Map<String, dynamic> json) {
    return AppSettings(
      appName: (json['app_name'] as String?)?.trim().isNotEmpty == true
          ? json['app_name'] as String
          : 'BaseAdmin',
      logoUrl: (json['app_logo'] as String?)?.isNotEmpty == true ? json['app_logo'] as String : null,
      faviconUrl: (json['app_favicon'] as String?)?.isNotEmpty == true ? json['app_favicon'] as String : null,
      primaryColor: _parseHexColor(json['primary_color'] as String?) ?? AppColors.defaultPrimary,
    );
  }

  static Color? _parseHexColor(String? hex) {
    if (hex == null || hex.isEmpty) return null;
    var value = hex.trim().replaceFirst('#', '');
    if (value.length == 6) value = 'FF$value';
    if (value.length != 8) return null;
    final parsed = int.tryParse(value, radix: 16);
    return parsed == null ? null : Color(parsed);
  }
}
