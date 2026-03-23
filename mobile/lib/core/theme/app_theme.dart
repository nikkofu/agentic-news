import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'app_colors.dart';

class AppTheme {
  static ThemeData get lightTheme {
    final textTheme = GoogleFonts.manropeTextTheme();
    return ThemeData(
      brightness: Brightness.light,
      primaryColor: AppColors.primary,
      scaffoldBackgroundColor: AppColors.background,
      colorScheme: const ColorScheme.light(
        primary: AppColors.primary,
        onPrimary: AppColors.onPrimary,
        primaryContainer: AppColors.primaryContainer,
        onPrimaryContainer: AppColors.onPrimaryContainer,
        secondary: AppColors.secondary,
        onSecondary: AppColors.onSecondary,
        secondaryContainer: AppColors.secondaryContainer,
        onSecondaryContainer: AppColors.onSecondaryContainer,
        surface: AppColors.surface,
        onSurface: AppColors.onSurface,
        background: AppColors.background,
        onBackground: AppColors.onBackground,
        error: Colors.redAccent,
        onError: Colors.white,
      ),
      textTheme: textTheme.copyWith(
        displayLarge: GoogleFonts.notoSerif(textStyle: textTheme.displayLarge, color: AppColors.primary),
        displayMedium: GoogleFonts.notoSerif(textStyle: textTheme.displayMedium, color: AppColors.primary),
        bodyLarge: textTheme.bodyLarge?.copyWith(color: AppColors.onSurface),
        bodyMedium: textTheme.bodyMedium?.copyWith(color: AppColors.onSurfaceVariant),
      ),
    );
  }

  static ThemeData get darkTheme {
    final textTheme = GoogleFonts.manropeTextTheme(ThemeData.dark().textTheme);
    return ThemeData(
      brightness: Brightness.dark,
      primaryColor: AppColors.darkPrimary,
      scaffoldBackgroundColor: AppColors.darkBackground,
      colorScheme: const ColorScheme.dark(
        primary: AppColors.darkPrimary,
        onPrimary: AppColors.darkOnPrimary,
        primaryContainer: AppColors.darkPrimaryContainer,
        onPrimaryContainer: AppColors.darkOnPrimaryContainer,
        secondary: AppColors.secondaryContainer,
        onSecondary: AppColors.onSecondaryContainer,
        surface: AppColors.darkSurface,
        onSurface: AppColors.darkOnSurface,
        background: AppColors.darkBackground,
        onBackground: AppColors.darkOnBackground,
        error: Colors.redAccent,
        onError: Colors.black,
      ),
      textTheme: textTheme.copyWith(
        displayLarge: GoogleFonts.notoSerif(textStyle: textTheme.displayLarge, color: AppColors.darkPrimary),
        displayMedium: GoogleFonts.notoSerif(textStyle: textTheme.displayMedium, color: AppColors.darkPrimary),
        bodyLarge: textTheme.bodyLarge?.copyWith(color: AppColors.darkOnSurface),
        bodyMedium: textTheme.bodyMedium?.copyWith(color: AppColors.darkOnSurfaceVariant),
      ),
    );
  }
}
