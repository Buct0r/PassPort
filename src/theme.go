package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CustomTheme struct{}

func (c *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 1, G: 22, B: 39, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 10, G: 30, B: 50, A: 220}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (c *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (c *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

type DarkTheme struct{}

func (c *DarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (c *DarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c *DarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (c *DarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

type LightTheme struct{}

func (c *LightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 240, G: 240, B: 240, A: 220}
	case theme.ColorNameButton:
		return color.RGBA{R: 230, G: 230, B: 230, A: 255}
	case theme.ColorNameHover:
		return color.RGBA{R: 200, G: 200, B: 200, A: 255}
	case theme.ColorNameForeground:
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0, G: 100, B: 200, A: 255}
	case theme.ColorNameError:
		return color.RGBA{R: 200, G: 0, B: 0, A: 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 200, G: 200, B: 200, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 0, G: 0, B: 0, A: 128}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (c *LightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (c *LightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (c *LightTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
