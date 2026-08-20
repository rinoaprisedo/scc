import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/theme/app_colors.dart';
import 'more_menu_sheet.dart';

/// Bottom-nav shell wrapping the authenticated area. Uses go_router's
/// StatefulNavigationShell so each routed tab keeps its own navigation
/// state; the trailing "More" tab isn't a route branch — it opens a sheet.
class MainShell extends ConsumerStatefulWidget {
  final StatefulNavigationShell navigationShell;

  const MainShell({super.key, required this.navigationShell});

  @override
  ConsumerState<MainShell> createState() => _MainShellState();
}

class _MainShellState extends ConsumerState<MainShell> {
  static const _tabs = [
    (icon: Icons.home_rounded, label: 'Home'),
    (icon: Icons.explore_rounded, label: 'Explore'),
    (icon: Icons.calendar_today_rounded, label: 'Calendar'),
    (icon: Icons.mail_rounded, label: 'Inbox'),
  ];
  static const _moreLabel = (icon: Icons.more_horiz_rounded, label: 'More');

  bool _navVisible = true;

  bool _onScrollNotification(ScrollNotification notification) {
    if (notification is UserScrollNotification && notification.direction != ScrollDirection.idle) {
      final hide = notification.direction == ScrollDirection.reverse;
      if (hide != !_navVisible) setState(() => _navVisible = !hide);
    }
    return false;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      extendBody: true,
      body: NotificationListener<ScrollNotification>(
        onNotification: _onScrollNotification,
        child: widget.navigationShell,
      ),
      bottomNavigationBar: AnimatedSlide(
        duration: const Duration(milliseconds: 500),
        curve: Curves.easeInOutCubic,
        offset: _navVisible ? Offset.zero : const Offset(0, 1.4),
        child: AnimatedOpacity(
          duration: const Duration(milliseconds: 450),
          curve: Curves.easeInOut,
          opacity: _navVisible ? 1 : 0,
          child: SafeArea(
            minimum: const EdgeInsets.fromLTRB(24, 0, 24, 16),
            child: _FloatingNavBar(
              currentIndex: widget.navigationShell.currentIndex,
              items: _tabs,
              moreItem: _moreLabel,
              onTapTab: (index) => widget.navigationShell.goBranch(
                index,
                initialLocation: index == widget.navigationShell.currentIndex,
              ),
              onTapMore: () => showMoreMenuSheet(context, ref),
            ),
          ),
        ),
      ),
    );
  }
}

class _FloatingNavBar extends StatelessWidget {
  final int currentIndex;
  final List<({IconData icon, String label})> items;
  final ({IconData icon, String label}) moreItem;
  final ValueChanged<int> onTapTab;
  final VoidCallback onTapMore;

  const _FloatingNavBar({
    required this.currentIndex,
    required this.items,
    required this.moreItem,
    required this.onTapTab,
    required this.onTapMore,
  });

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;

    return Container(
      height: 64,
      padding: const EdgeInsets.symmetric(horizontal: 6),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(28),
        border: Border.all(color: AppColors.border),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.08),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: [
          for (var i = 0; i < items.length; i++)
            Expanded(
              child: _NavIcon(
                icon: items[i].icon,
                label: items[i].label,
                selected: i == currentIndex,
                color: primary,
                onTap: () => onTapTab(i),
              ),
            ),
          Expanded(
            child: _NavIcon(
              icon: moreItem.icon,
              label: moreItem.label,
              selected: false,
              color: primary,
              onTap: onTapMore,
            ),
          ),
        ],
      ),
    );
  }
}

class _NavIcon extends StatelessWidget {
  final IconData icon;
  final String label;
  final bool selected;
  final Color color;
  final VoidCallback onTap;

  const _NavIcon({
    required this.icon,
    required this.label,
    required this.selected,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        mainAxisSize: MainAxisSize.min,
        children: [
          AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            curve: Curves.easeOut,
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: selected ? color : Colors.transparent,
              shape: BoxShape.circle,
            ),
            child: Icon(
              icon,
              size: 19,
              color: selected ? Colors.white : AppColors.textSecondary,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              fontWeight: selected ? FontWeight.w600 : FontWeight.w500,
              color: selected ? color : AppColors.textSecondary,
            ),
          ),
        ],
      ),
    );
  }
}
