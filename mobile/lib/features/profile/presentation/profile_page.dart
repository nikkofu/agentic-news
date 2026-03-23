import 'package:flutter/material.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(title: Text('Settings & Profile', style: theme.textTheme.labelLarge?.copyWith(letterSpacing: 2.0))),
      body: ListView(
        children: [
          Container(
            padding: const EdgeInsets.all(24),
            color: theme.colorScheme.surface,
            child: Row(
              children: [
                CircleAvatar(radius: 40, backgroundColor: theme.colorScheme.primaryContainer, child: Icon(Icons.person, size: 40, color: theme.colorScheme.onPrimaryContainer)),
                const SizedBox(width: 24),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Sanctuary Architect', style: theme.textTheme.titleLarge?.copyWith(fontWeight: FontWeight.bold)),
                    const SizedBox(height: 4),
                    Text('patron@digital-sanctuary', style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.primary)),
                  ],
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          ListTile(title: const Text('Focus Goal (Mins)'), trailing: const Text('120'), leading: const Icon(Icons.timer)),
          SwitchListTile(title: const Text('Dark Mode Lock'), value: true, onChanged: (v) {}, secondary: const Icon(Icons.dark_mode)),
          SwitchListTile(title: const Text('Oracle Daily Push'), subtitle: const Text('Receive end-of-day synthesis'), value: true, onChanged: (v) {}, secondary: const Icon(Icons.notifications)),
          ListTile(title: const Text('Local Vault Path'), subtitle: const Text('/Users/patron/.openclaw/vault'), leading: const Icon(Icons.folder)),
          const Divider(),
          ListTile(title: Text('Sign Out', style: TextStyle(color: theme.colorScheme.error)), leading: Icon(Icons.logout, color: theme.colorScheme.error), onTap: () {}),
        ],
      ),
    );
  }
}
