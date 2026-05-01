//go:build windows
// +build windows

package main

import (
	"syscall"
)

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procFreeConsole = kernel32.NewProc("FreeConsole")
)

// hideConsole hides the console window on Windows (GUI mode)
func hideConsole() {
	procFreeConsole.Call()
}
