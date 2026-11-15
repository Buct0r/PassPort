package theme

import (
	"os"
	"path/filepath"
	"runtime"
)

// platformConfigDir returns the per-platform directory where the configuration
// file is stored.
func platformConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "PassPort")
		}
	case "linux":
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".config", "passport")
		}
	case "darwin":
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, "Library", "Application Support", "PassPort")
		}
	}
	return ""
}

// ConfigPath returns the path of config.json. An existing file is preferred, in
// the order: executable directory, working directory, platform config
// directory. When no file exists yet the platform config directory is returned
// as the location to create it in.
func ConfigPath() string {
	var candidates []string

	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "config.json"))
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "config.json"))
	}
	if dir := platformConfigDir(); dir != "" {
		candidates = append(candidates, filepath.Join(dir, "config.json"))
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	if len(candidates) > 0 {
		return candidates[len(candidates)-1]
	}
	return ""
}
