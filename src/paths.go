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

// getConfigPath returns the platform-specific path for configuration files
func getConfigPath() string {
	// Check executable directory first
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		configPath := filepath.Join(exeDir, "config.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Check current working directory
	if cwd, err := os.Getwd(); err == nil {
		configPath := filepath.Join(cwd, "config.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Check platform-specific config directories
	var configDir string
	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata != "" {
			configDir = filepath.Join(appdata, "PassPort")
		}

	case "linux":
		home := os.Getenv("HOME")
		if home != "" {
			configDir = filepath.Join(home, ".config", "passport")
		}

	case "darwin":
		home := os.Getenv("HOME")
		if home != "" {
			configDir = filepath.Join(home, "Library", "Application Support", "PassPort")
		}
	}

	if configDir != "" {
		configPath := filepath.Join(configDir, "config.json")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// Not found anywhere
	return ""
}
