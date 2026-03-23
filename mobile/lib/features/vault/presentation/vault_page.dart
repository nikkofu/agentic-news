import 'package:flutter/material.dart';

class VaultPage extends StatelessWidget {
  const VaultPage({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      backgroundColor: theme.colorScheme.background,
      appBar: AppBar(
        title: Text('The Knowledge Vault', style: theme.textTheme.labelLarge?.copyWith(letterSpacing: 2.0, fontWeight: FontWeight.bold)),
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () {},
          )
        ],
      ),
      body: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 8.0),
              child: Text(
                'YOUR COGNITIVE NETWORK',
                style: theme.textTheme.labelSmall?.copyWith(letterSpacing: 2.0, color: theme.colorScheme.onSurfaceVariant),
              ),
            ),
            const SizedBox(height: 24),
            Expanded(
              child: Container(
                margin: const EdgeInsets.symmetric(horizontal: 16.0),
                decoration: BoxDecoration(
                  gradient: RadialGradient(
                    colors: [
                      theme.colorScheme.primaryContainer.withOpacity(0.4),
                      theme.colorScheme.surface,
                    ],
                    radius: 1.5,
                  ),
                  borderRadius: BorderRadius.circular(32),
                  border: Border.all(color: theme.colorScheme.outline.withOpacity(0.1)),
                ),
                child: Stack(
                  children: [
                    _buildMockNode(context, 'Stoicism', 0.2, 0.3, true),
                    _buildMockNode(context, 'LLM Alignment', 0.6, 0.2, false),
                    _buildMockNode(context, 'Deep Work', 0.4, 0.6, true),
                    _buildMockNode(context, 'Systems Thinking', 0.7, 0.7, false),
                    _buildMockNode(context, 'Entropy', 0.1, 0.8, false),
                    Center(
                      child: Container(
                        padding: const EdgeInsets.all(24),
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: theme.colorScheme.primary.withOpacity(0.1),
                          border: Border.all(color: theme.colorScheme.primary.withOpacity(0.3)),
                        ),
                        child: Icon(Icons.hub, size: 48, color: theme.colorScheme.primary),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24.0),
              child: Text('Recent Archives', style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.bold)),
            ),
            const SizedBox(height: 16),
            Expanded(
              flex: 0,
              child: SizedBox(
                height: 120,
                child: ListView(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 24.0),
                  children: [
                    _buildArchiveCard(context, 'The Architecture of Silence', '2 hours ago'),
                    const SizedBox(width: 16),
                    _buildArchiveCard(context, 'Moral Ghost in the Machine', 'Yesterday'),
                    const SizedBox(width: 16),
                    _buildArchiveCard(context, 'Git Deployment Postmortem', '3 days ago'),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildMockNode(BuildContext context, String label, double x, double y, bool primary) {
    final theme = Theme.of(context);
    return Align(
      alignment: FractionalOffset(x, y),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: primary ? theme.colorScheme.primary : theme.colorScheme.surfaceVariant,
          borderRadius: BorderRadius.circular(20),
          boxShadow: [
            if (primary) BoxShadow(color: theme.colorScheme.primary.withOpacity(0.3), blurRadius: 10, spreadRadius: 2)
          ],
        ),
        child: Text(
          label,
          style: theme.textTheme.labelMedium?.copyWith(
            color: primary ? theme.colorScheme.onPrimary : theme.colorScheme.onSurfaceVariant,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );
  }

  Widget _buildArchiveCard(BuildContext context, String title, String time) {
    final theme = Theme.of(context);
    return Container(
      width: 200,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(title, style: theme.textTheme.labelLarge?.copyWith(fontWeight: FontWeight.bold), maxLines: 2, overflow: TextOverflow.ellipsis),
          Text(time, style: theme.textTheme.labelSmall?.copyWith(color: theme.colorScheme.outline)),
        ],
      ),
    );
  }
}
