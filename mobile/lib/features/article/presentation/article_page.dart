import 'package:flutter/material.dart';

class ArticlePage extends StatelessWidget {
  final String id;
  const ArticlePage({super.key, required this.id});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 300.0,
            floating: false,
            pinned: true,
            backgroundColor: theme.colorScheme.background,
            flexibleSpace: FlexibleSpaceBar(
              title: Text('The Architecture of Silence', style: theme.textTheme.titleMedium?.copyWith(
                color: Colors.white, shadows: [Shadow(color: Colors.black54, blurRadius: 10)]
              )),
              background: Stack(
                fit: StackFit.expand,
                children: [
                  Image.network(
                    'https://images.unsplash.com/photo-1541963463532-d68292c34b19?q=80&w=2500&auto=format&fit=crop',
                    fit: BoxFit.cover,
                  ),
                  const DecoratedBox(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topCenter,
                        end: Alignment.bottomCenter,
                        colors: [Colors.transparent, Colors.black87],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 32.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Chip(
                        label: const Text('DEEP FOCUS', style: TextStyle(fontSize: 10, letterSpacing: 1.5, fontWeight: FontWeight.bold)),
                        backgroundColor: theme.colorScheme.secondaryContainer.withOpacity(0.5),
                        side: BorderSide.none,
                      ),
                      const SizedBox(width: 12),
                      Text('12 MIN READ', style: theme.textTheme.labelSmall?.copyWith(color: theme.colorScheme.outline, letterSpacing: 1.0)),
                    ],
                  ),
                  const SizedBox(height: 24),
                  Text(
                    'In an era governed by noise and constant context switching, silence is no longer an absence of sound but a deliberate constructed environment.',
                    style: theme.textTheme.bodyLarge?.copyWith(height: 1.8, fontSize: 18),
                  ),
                  const SizedBox(height: 32),
                  Text(
                    '1. The Illusion of Productivity',
                    style: theme.textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold, color: theme.colorScheme.primary),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'We stack tools upon tools—pagers, monitors, notification sinks. Each layer abstracting us further from the raw material of thought. The true scholar builds an atrium, a moat against the entropy of real-time alerts.',
                    style: theme.textTheme.bodyMedium?.copyWith(height: 1.8, fontSize: 16, color: theme.colorScheme.onSurfaceVariant),
                  ),
                  const SizedBox(height: 64), // Scroll padding
                ],
              ),
            ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {},
        icon: const Icon(Icons.bookmark_add_outlined),
        label: const Text('Send to Vault'),
        backgroundColor: theme.colorScheme.primary,
        foregroundColor: theme.colorScheme.onPrimary,
      ),
    );
  }
}
