package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	sharedtheme "github.com/Buct0r/PassPort/internal/theme"
)

// UserTheme adapts a shared theme palette (a map of hex colors) to Fyne's
// fyne.Theme interface. Any color that is absent from the palette falls back
// to Fyne's default theme, so a theme only needs to define what it changes.
type UserTheme struct {
	colors sharedtheme.Theme
}

// NewUserTheme wraps a shared palette in a Fyne theme.
func NewUserTheme(t sharedtheme.Theme) *UserTheme {
	if t == nil {
		t = sharedtheme.Theme{}
	}
	return &UserTheme{colors: t}
}

// Color resolves a Fyne color name against the palette, falling back to the
// Fyne default when the key is missing or invalid.
func (u *UserTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if c, ok := u.colors.Color(string(name)); ok {
		return c
	}
	return theme.DefaultTheme().Color(name, variant)
}

// CardBackground returns the custom password-card background color, falling
// back to the app background when the theme does not define one.
func (u *UserTheme) CardBackground() color.Color {
	if c, ok := u.colors.Color(sharedtheme.CardBackgroundKey); ok {
		return c
	}
	return theme.DefaultTheme().Color(theme.ColorNameBackground, theme.VariantDark)
}

// Foreground is the primary text color.
func (u *UserTheme) Foreground() color.Color {
	return u.Color(theme.ColorNameForeground, theme.VariantDark)
}

// Muted is a lower-emphasis text color, used for captions and subtitles.
func (u *UserTheme) Muted() color.Color {
	return u.Color(theme.ColorNamePlaceHolder, theme.VariantDark)
}

// Primary is the accent color.
func (u *UserTheme) Primary() color.Color {
	return u.Color(theme.ColorNamePrimary, theme.VariantDark)
}

// Separator is the color for divider lines and borders.
func (u *UserTheme) Separator() color.Color {
	return u.Color(theme.ColorNameSeparator, theme.VariantDark)
}

// Icon delegates to Fyne's default icon set.
func (u *UserTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Font delegates to Fyne's default font.
func (u *UserTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Size delegates to Fyne's default sizing.
func (u *UserTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
