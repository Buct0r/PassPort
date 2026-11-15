package main

import (
	sharedtheme "github.com/Buct0r/PassPort/internal/theme"
)

// cliColors holds the terminal colors loaded from config.json. It defaults to
// the built-in palette and is refreshed by initCLITheme.
var cliColors = sharedtheme.DefaultConfig().EffectiveCLI()

// initCLITheme loads the CLI color configuration from config.json.
func initCLITheme() {
	cliColors = sharedtheme.LoadConfig().EffectiveCLI()
}

// accentColor is the escape sequence for branding and prompts.
func accentColor() string {
	return sharedtheme.ANSIColor(cliColors.Accent, 36)
}

// errorColor is the escape sequence for errors and warnings.
func errorColor() string {
	return sharedtheme.ANSIColor(cliColors.Error, 31)
}

// successColor is the escape sequence for success messages.
func successColor() string {
	return sharedtheme.ANSIColor(cliColors.Success, 32)
}

// warningColor is the escape sequence for warnings.
func warningColor() string {
	return sharedtheme.ANSIColor(cliColors.Warning, 33)
}

// resetColor restores the terminal's default colors.
func resetColor() string {
	return sharedtheme.Reset()
}
