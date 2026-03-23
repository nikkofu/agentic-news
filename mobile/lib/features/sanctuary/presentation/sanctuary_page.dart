import 'package:flutter/material.dart';

class SanctuaryPage extends StatelessWidget {
  const SanctuaryPage({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text('The Oracle\'s Mirror', style: theme.textTheme.labelLarge?.copyWith(letterSpacing: 2.0, fontWeight: FontWeight.bold)),
        centerTitle: false,
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.symmetric(horizontal: 24.0, vertical: 16.0),
          children: [
            Text(
              'Daily Reflection',
              style: theme.textTheme.displayMedium?.copyWith(
                fontStyle: FontStyle.italic,
                fontWeight: FontWeight.w500,
                color: theme.colorScheme.primary,
              ),
            ),
            const SizedBox(height: 16),
            Text(
              'Synthesize your cognitive footprint for the day. Confess your friction across the endless scroll of time.',
              style: theme.textTheme.bodyLarge?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
                height: 1.5,
              ),
            ),
            const SizedBox(height: 40),
            Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                color: theme.colorScheme.surfaceContainerHighest.withOpacity(0.5),
                borderRadius: BorderRadius.circular(24),
                border: Border.all(color: theme.colorScheme.outline.withOpacity(0.1)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.psychology, color: theme.colorScheme.primary),
                      const SizedBox(width: 12),
                      Text('The Crucible of Friction', style: theme.textTheme.labelLarge?.copyWith(fontWeight: FontWeight.bold, letterSpacing: 1.0)),
                    ],
                  ),
                  const SizedBox(height: 20),
                  TextField(
                    maxLines: 5,
                    style: theme.textTheme.bodyMedium,
                    decoration: InputDecoration(
                      hintText: 'What systemic or architectural blockages drained your focus today?',
                      hintStyle: TextStyle(color: theme.colorScheme.onSurfaceVariant.withOpacity(0.5)),
                      filled: true,
                      fillColor: theme.colorScheme.surface,
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(16), borderSide: BorderSide.none),
                      contentPadding: const EdgeInsets.all(20),
                    ),
                  ),
                  const SizedBox(height: 24),
                  FilledButton.icon(
                    onPressed: () {},
                    icon: const Icon(Icons.auto_awesome),
                    label: const Text('Consult the Oracle'),
                    style: FilledButton.styleFrom(
                      minimumSize: const Size(double.infinity, 60),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      backgroundColor: theme.colorScheme.primary,
                      foregroundColor: theme.colorScheme.onPrimary,
                      textStyle: theme.textTheme.labelLarge?.copyWith(fontWeight: FontWeight.bold, letterSpacing: 1.2),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 48),
            // Placeholder for Oracle's insights (loaded dynamically later)
            Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                gradient: LinearBinding(theme),
                borderRadius: BorderRadius.circular(24),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('ORACLE\'S WHISPER', style: theme.textTheme.labelSmall?.copyWith(letterSpacing: 2.0, color: theme.colorScheme.onPrimary.withOpacity(0.7))),
                  const SizedBox(height: 12),
                  Text(
                    '"Resistance is exactly where your mind is mapping the unknown. You fought Git permissions today to protect your deployment immutability tomorrow."',
                    style: theme.textTheme.bodyLarge?.copyWith(
                      fontStyle: FontStyle.italic,
                      color: theme.colorScheme.onPrimary,
                      height: 1.6,
                    ),
                  )
                ],
              ),
            )
          ],
        ),
      ),
    );
  }

  LinearGradient LinearBinding(ThemeData theme) {
     return LinearGradient(
       colors: [
         theme.colorScheme.primary,
         theme.colorScheme.primaryContainer,
       ],
       begin: Alignment.topLeft,
       end: Alignment.bottomRight,
     );
  }
}
