//go:build darwin

package main

import (
	"fmt"
	"os"
)

func gui() {
	if os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" {
		fmt.Println("No display detected. Use --cli for CLI mode.")
		return
	}
	runFyneGUI()
}
