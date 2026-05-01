//go:build darwin
// +build darwin

package main

import "fmt"

// gui is a no-op on macOS due to Fyne OpenGL dependency limitations
// Users should use CLI mode on macOS
func gui() {
	fmt.Println("Note: GUI mode is not available on macOS in this version due to Fyne OpenGL dependency constraints.")
	fmt.Println("Please use CLI mode with the -cli flag instead.")
}
