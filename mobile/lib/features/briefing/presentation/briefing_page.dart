import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../data/repositories/briefing_repository.dart';

class BriefingPage extends ConsumerWidget {
  const BriefingPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final briefingAsync = ref.watch(dailyBriefingProvider);

    return Scaffold(
      body: SafeArea(
        child: briefingAsync.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (err, stack) => Center(child: Text('Error loading sanctuary: $err')),
          data: (briefing) {
            final insight = briefing.topInsight;
            return Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Focus of the Day',
                    style: theme.textTheme.labelSmall?.copyWith(
                      letterSpacing: 3.0,
                      fontWeight: FontWeight.w800,
                      color: theme.colorScheme.secondary,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    insight?.title ?? 'No Insight Title',
                    style: theme.textTheme.displayMedium?.copyWith(
                      fontStyle: FontStyle.italic,
                      fontWeight: FontWeight.w500,
                      height: 1.15,
                    ),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    insight?.summary ?? 'No summary available.',
                    style: theme.textTheme.bodyLarge?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                      height: 1.5,
                    ),
                  ),
                  const SizedBox(height: 24),
                  FilledButton.icon(
                    onPressed: () => context.push('/article/1'),
                    icon: const Icon(Icons.arrow_forward),
                    label: const Text('Begin Deep Dive'),
                    style: FilledButton.styleFrom(
                      padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 18),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      textStyle: theme.textTheme.labelLarge?.copyWith(fontWeight: FontWeight.bold)
                    ),
                  ),
                  const SizedBox(height: 32),
                  if (insight?.imageUrl != null)
                    Expanded(
                      child: Container(
                        width: double.infinity,
                        decoration: BoxDecoration(
                          color: theme.colorScheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(24),
                          image: DecorationImage(
                            image: NetworkImage(insight!.imageUrl),
                            fit: BoxFit.cover,
                            colorFilter: ColorFilter.mode(Colors.black.withOpacity(0.2), BlendMode.darken),
                          )
                        ),
                      ),
                    )
                  else 
                    Expanded(
                      child: Container(
                        width: double.infinity,
                        decoration: BoxDecoration(
                          color: theme.colorScheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(24),
                        ),
                        child: const Center(
                          child: Icon(Icons.auto_stories, size: 48, color: Colors.white54),
                        ),
                      ),
                    )
                ],
              ),
            );
          }
        )
      ),
    );
  }
}
