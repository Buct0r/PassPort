package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// getMasterPasswordPath returns the platform-specific path for storing the master password
func getMasterPasswordPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return "", fmt.Errorf("cannot determine AppData directory")
		}
		configDir = filepath.Join(appdata, ".passport")

	case "linux":
		home := os.Getenv("HOME")
		if home == "" {
			usr, err := user.Current()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory")
			}
			home = usr.HomeDir
		}
		configDir = filepath.Join(home, ".config", "passport")

	case "darwin":
		home := os.Getenv("HOME")
		if home == "" {
			usr, err := user.Current()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory")
			}
			home = usr.HomeDir
		}
		configDir = filepath.Join(home, "Library", "Application Support", "PassPort")

	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "master.key"), nil
}

// getSecretFilePath returns the platform-specific path for storing encrypted passwords
func getSecretFilePath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return "", fmt.Errorf("cannot determine AppData directory")
		}
		configDir = filepath.Join(appdata, ".passport")

	case "linux":
		home := os.Getenv("HOME")
		if home == "" {
			usr, err := user.Current()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory")
			}
			home = usr.HomeDir
		}
		configDir = filepath.Join(home, ".config", "passport")

	case "darwin":
		home := os.Getenv("HOME")
		if home == "" {
			usr, err := user.Current()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory")
			}
			home = usr.HomeDir
		}
		configDir = filepath.Join(home, "Library", "Application Support", "PassPort")

	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "secret"), nil
}
