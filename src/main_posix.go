//go:build !windows
// +build !windows

package main

// hideConsole is a no-op on non-Windows platforms
func hideConsole() {
	// Console hiding is only needed on Windows
}
