import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../features/app/presentation/app_layout.dart';
import '../../features/briefing/presentation/briefing_page.dart';
import '../../features/sanctuary/presentation/sanctuary_page.dart';
import '../../features/article/presentation/article_page.dart';
import '../../features/vault/presentation/vault_page.dart';
import '../../features/onboarding/presentation/onboarding_page.dart';
import '../../features/profile/presentation/profile_page.dart';

final rootNavigatorKey = GlobalKey<NavigatorState>();
final shellNavigatorKey = GlobalKey<NavigatorState>();

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    navigatorKey: rootNavigatorKey,
    initialLocation: '/',
    routes: [
      ShellRoute(
        navigatorKey: shellNavigatorKey,
        builder: (context, state, child) {
          return AppLayout(child: child, currentPath: state.fullPath ?? '/');
        },
        routes: [
          GoRoute(
            path: '/',
            builder: (context, state) => const BriefingPage(),
          ),
          GoRoute(
            path: '/sanctuary',
            builder: (context, state) => const SanctuaryPage(),
          ),
          GoRoute(
            path: '/vault',
            builder: (context, state) => const VaultPage(),
          ),
        ],
      ),
      GoRoute(
        path: '/article/:id',
        parentNavigatorKey: rootNavigatorKey,
        builder: (context, state) => ArticlePage(id: state.pathParameters['id']!),
      ),
      GoRoute(
        path: '/onboarding',
        parentNavigatorKey: rootNavigatorKey,
        builder: (context, state) => const OnboardingPage(),
      ),
      GoRoute(
        path: '/profile',
        parentNavigatorKey: rootNavigatorKey,
        builder: (context, state) => const ProfilePage(),
      ),
    ],
  );
});
