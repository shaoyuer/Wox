import 'dart:ui';
import 'package:flutter/material.dart' show HSLColor;
import 'package:from_css_color/from_css_color.dart';
import 'package:wox/utils/wox_theme_util.dart';

Color getThemeActiveBackgroundColor() {
  return fromCssColor(WoxThemeUtil.instance.currentTheme.value.actionItemActiveBackgroundColor);
}

Color getThemeTextColor() {
  return fromCssColor(WoxThemeUtil.instance.currentTheme.value.resultItemTitleColor);
}

Color getThemeSubTextColor() {
  return fromCssColor(WoxThemeUtil.instance.currentTheme.value.resultItemSubTitleColor);
}

Color getThemeBackgroundColor() {
  return fromCssColor(WoxThemeUtil.instance.currentTheme.value.appBackgroundColor);
}

Color getThemeDividerColor() {
  return fromCssColor(WoxThemeUtil.instance.currentTheme.value.previewSplitLineColor);
}

Color getThemeActionItemActiveColor() {
  return fromCssColor(WoxThemeUtil.instance.currentTheme.value.actionItemActiveFontColor);
}

Color getThemeCardBackgroundColor() {
  Color baseColor = getThemeBackgroundColor();
  bool isDarkTheme = baseColor.computeLuminance() < 0.5;
  if (isDarkTheme) {
    return Color.fromRGBO(
      (baseColor.r + 20 > 255 ? 255 : baseColor.r + 20).toInt(),
      (baseColor.g + 20 > 255 ? 255 : baseColor.g + 20).toInt(),
      (baseColor.b + 20 > 255 ? 255 : baseColor.b + 20).toInt(),
      1.0,
    );
  } else {
    return Color.fromRGBO(
      (baseColor.r - 20 < 0 ? 0 : baseColor.r - 20).toInt(),
      (baseColor.g - 20 < 0 ? 0 : baseColor.g - 20).toInt(),
      (baseColor.b - 20 < 0 ? 0 : baseColor.b - 20).toInt(),
      1.0,
    );
  }
}

extension ColorsExtension on Color {
  /// Returns a darker version of the color.
  ///
  /// [weight] is a value between 0 and 100 that determines how much darker the color should be.
  /// A weight of 0 returns the original color, while a weight of 100 returns black.
  Color darker([int weight = 10]) {
    weight = weight.clamp(0, 100);
    final hslColor = HSLColor.fromColor(this);
    final newLightness = (hslColor.lightness * (100 - weight) / 100).clamp(0.0, 1.0);
    return hslColor.withLightness(newLightness).toColor();
  }

  /// Returns a lighter version of the color.
  ///
  /// [weight] is a value between 0 and 100 that determines how much lighter the color should be.
  /// A weight of 0 returns the original color, while a weight of 100 returns white.
  Color lighter([int weight = 10]) {
    weight = weight.clamp(0, 100);
    final hslColor = HSLColor.fromColor(this);
    final newLightness = (hslColor.lightness + (1 - hslColor.lightness) * weight / 100).clamp(0.0, 1.0);
    return hslColor.withLightness(newLightness).toColor();
  }
}
