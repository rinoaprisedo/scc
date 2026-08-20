import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/router/routes.dart';
import '../core/theme/app_colors.dart';
import '../features/auth/presentation/auth_provider.dart';

class _MoreMenuItem {
  final IconData icon;
  final String label;
  final VoidCallback? Function(BuildContext context)? action;

  const _MoreMenuItem(this.icon, this.label, [this.action]);
}

List<_MoreMenuItem> _buildItems(WidgetRef ref) => [
  _MoreMenuItem(
    Icons.person_rounded,
    'Profile',
    (context) => () {
      Navigator.of(context).pop();
      context.push(Routes.profile);
    },
  ),
  const _MoreMenuItem(Icons.notifications_rounded, 'Notifications'),
  const _MoreMenuItem(Icons.bookmark_rounded, 'Saved'),
  const _MoreMenuItem(Icons.settings_rounded, 'Settings'),
  const _MoreMenuItem(Icons.language_rounded, 'Language'),
  const _MoreMenuItem(Icons.dark_mode_rounded, 'Appearance'),
  const _MoreMenuItem(Icons.payment_rounded, 'Payments'),
  const _MoreMenuItem(Icons.card_giftcard_rounded, 'Promotions'),
  const _MoreMenuItem(Icons.support_agent_rounded, 'Help'),
  const _MoreMenuItem(Icons.privacy_tip_rounded, 'Privacy'),
  const _MoreMenuItem(Icons.description_rounded, 'Terms'),
  const _MoreMenuItem(Icons.info_rounded, 'About'),
  const _MoreMenuItem(Icons.star_rounded, 'Rate App'),
  const _MoreMenuItem(Icons.share_rounded, 'Share App'),
  const _MoreMenuItem(Icons.group_add_rounded, 'Invite'),
  _MoreMenuItem(
    Icons.logout_rounded,
    'Logout',
    (context) => () {
      Navigator.of(context).pop();
      ref.read(authProvider.notifier).logout();
    },
  ),
];

/// Placeholder menu grid surfaced from the bottom nav's "More" tab — styled
/// like the floating nav bar itself (rounded white card, circle icon chips)
/// and anchored just above it, rather than a full-height sheet. Swap each
/// dummy item for a real route/action as those features get built.
Future<void> showMoreMenuSheet(BuildContext context, WidgetRef ref) {
  final primary = Theme.of(context).colorScheme.primary;
  final items = _buildItems(ref);

  return showGeneralDialog(
    context: context,
    barrierDismissible: true,
    barrierLabel: 'More menu',
    barrierColor: Colors.black.withValues(alpha: 0.25),
    transitionDuration: const Duration(milliseconds: 220),
    transitionBuilder: (context, animation, secondaryAnimation, child) {
      final curved = CurvedAnimation(
        parent: animation,
        curve: Curves.easeOutCubic,
      );
      return FadeTransition(
        opacity: curved,
        child: SlideTransition(
          position: Tween(
            begin: const Offset(0, 0.08),
            end: Offset.zero,
          ).animate(curved),
          child: child,
        ),
      );
    },
    pageBuilder: (context, animation, secondaryAnimation) {
      return SafeArea(
        // 64 (nav bar height) + 16 (its own bottom margin) + 12 (gap) so the
        // popup sits just above the floating nav bar, not over it.
        minimum: const EdgeInsets.fromLTRB(24, 0, 24, 92),
        child: Align(
          alignment: Alignment.bottomCenter,
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxHeight: 380),
            child: Container(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(28),
                border: Border.all(color: AppColors.border),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.12),
                    blurRadius: 24,
                    offset: const Offset(0, 12),
                  ),
                ],
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Flexible(
                    child: ScrollConfiguration(
                      behavior: ScrollConfiguration.of(
                        context,
                      ).copyWith(scrollbars: false),
                      child: GridView.builder(
                        shrinkWrap: true,
                        padding: const EdgeInsets.only(bottom: 8),
                        itemCount: items.length,
                        gridDelegate:
                            const SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: 4,
                              mainAxisSpacing: 8,
                              crossAxisSpacing: 4,
                              childAspectRatio: 0.8,
                            ),
                        itemBuilder: (context, index) {
                          final item = items[index];
                          final isLogout = item.label == 'Logout';
                          return _MoreGridTile(
                            icon: item.icon,
                            label: item.label,
                            color: isLogout ? AppColors.danger : primary,
                            onTap:
                                item.action?.call(context) ??
                                () {
                                  Navigator.of(context).pop();
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    SnackBar(
                                      content: Text(
                                        '${item.label} — coming soon',
                                      ),
                                    ),
                                  );
                                },
                          );
                        },
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      );
    },
  );
}

/// Same visual language as the bottom nav's icon chips (_NavIcon in
/// main_shell.dart): a tinted circle behind the icon, label below.
class _MoreGridTile extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  const _MoreGridTile({
    required this.icon,
    required this.label,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: color.withValues(alpha: 0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(icon, size: 20, color: color),
          ),
          const SizedBox(height: 6),
          Text(
            label,
            textAlign: TextAlign.center,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              fontSize: 10,
              fontWeight: FontWeight.w500,
              color: AppColors.textPrimary,
            ),
          ),
        ],
      ),
    );
  }
}
