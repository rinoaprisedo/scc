import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/widgets/floating_title_chip.dart';

class _Article {
  final String title;
  final String excerpt;
  final String category;
  final String readTime;
  final Color color;

  const _Article(
    this.title,
    this.excerpt,
    this.category,
    this.readTime,
    this.color,
  );
}

const _articles = [
  _Article(
    'Getting Started with Flutter',
    'A beginner-friendly walkthrough of widgets, state, and the framework\'s core building blocks.',
    'Tutorial',
    '5 min read',
    Color(0xFF3B82F6),
  ),
  _Article(
    'Designing Clean, Minimal UIs',
    'Why whitespace and restraint often make a bigger impact than adding more visual elements.',
    'Design',
    '4 min read',
    Color(0xFFA855F7),
  ),
  _Article(
    'State Management Compared',
    'A practical look at Riverpod, Bloc, and Provider — trade-offs for teams of different sizes.',
    'Engineering',
    '8 min read',
    Color(0xFF22C55E),
  ),
  _Article(
    'Shipping Your First Release',
    'A checklist for App Store and Play Store submissions, from signing to release notes.',
    'Guide',
    '6 min read',
    Color(0xFFF97316),
  ),
  _Article(
    'API Sessions vs. Tokens',
    'Cookie-based sessions and bearer tokens solve the same problem differently — here\'s when to use each.',
    'Backend',
    '7 min read',
    Color(0xFF06B6D4),
  ),
  _Article(
    'Accessible Color Palettes',
    'Choosing a primary color that still passes contrast checks across light and dark themes.',
    'Design',
    '3 min read',
    Color(0xFFEF4444),
  ),
  _Article(
    'Testing Widgets the Right Way',
    'Golden tests, widget tests, and when integration tests actually earn their cost.',
    'Engineering',
    '9 min read',
    Color(0xFF3B82F6),
  ),
  _Article(
    'Offline-First Mobile Apps',
    'Patterns for caching, syncing, and handling conflicts when connectivity drops.',
    'Architecture',
    '10 min read',
    Color(0xFFA855F7),
  ),
  _Article(
    'Animations That Feel Native',
    'Matching platform motion curves so custom transitions don\'t feel out of place.',
    'Design',
    '5 min read',
    Color(0xFF22C55E),
  ),
  _Article(
    'Securing User Sessions',
    'CSRF, httpOnly cookies, and why double-submit tokens still matter in 2026.',
    'Security',
    '6 min read',
    Color(0xFFF97316),
  ),
];

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          ScrollConfiguration(
            behavior: ScrollConfiguration.of(
              context,
            ).copyWith(scrollbars: false),
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(16, 44, 16, 96),
              itemCount: _articles.length,
              separatorBuilder: (context, index) => const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final article = _articles[index];
                return Container(
                  padding: const EdgeInsets.all(14),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: AppColors.border),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8,
                              vertical: 3,
                            ),
                            decoration: BoxDecoration(
                              color: article.color.withValues(alpha: 0.1),
                              borderRadius: BorderRadius.circular(20),
                            ),
                            child: Text(
                              article.category,
                              style: TextStyle(
                                fontSize: 10,
                                fontWeight: FontWeight.w700,
                                color: article.color,
                              ),
                            ),
                          ),
                          const Spacer(),
                          Text(
                            article.readTime,
                            style: const TextStyle(
                              fontSize: 11,
                              color: AppColors.textSecondary,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      Text(
                        article.title,
                        style: const TextStyle(
                          fontSize: 14.5,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        article.excerpt,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          fontSize: 12.5,
                          color: AppColors.textSecondary,
                          height: 1.4,
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
          const FloatingTitleChip(title: 'Home', icon: Icons.home_rounded),
        ],
      ),
    );
  }
}
