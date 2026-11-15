package theme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// CLIColors holds the terminal interface colors, expressed as hex strings.
type CLIColors struct {
	Accent  string `json:"accent"`
	Error   string `json:"error"`
	Success string `json:"success"`
	Warning string `json:"warning"`
}

// Config is the parsed contents of config.json.
type Config struct {
	Theme  string           `json:"theme"`
	Themes map[string]Theme `json:"themes"`
	CLI    CLIColors        `json:"cli"`
}

// DefaultConfig returns a configuration populated with every built-in theme
// and the default CLI colors.
func DefaultConfig() *Config {
	return &Config{
		Theme:  DefaultThemeName,
		Themes: DefaultThemes(),
		CLI:    defaultCLIColors,
	}
}

// LoadConfig reads config.json from ConfigPath(). When the file is missing it
// is created from DefaultConfig. Parse errors yield a default configuration so
// the application always starts.
func LoadConfig() *Config {
	cfg, _ := LoadConfigFrom(ConfigPath())
	return cfg
}

// LoadConfigFrom reads and normalises a configuration file at path. A missing
// file is created from DefaultConfig. The returned configuration is always
// safe to use; the error is informational.
func LoadConfigFrom(path string) (*Config, error) {
	if path == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		cfg := DefaultConfig()
		return cfg, SaveConfigTo(path, cfg)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("invalid config: %w", err)
	}

	migrated := normalise(&cfg)
	if migrated {
		_ = SaveConfigTo(path, &cfg)
	}
	return &cfg, nil
}

// normalise fills in missing fields (migrating older configs) and reports
// whether anything was changed.
func normalise(cfg *Config) bool {
	changed := false

	if len(cfg.Themes) == 0 {
		cfg.Themes = DefaultThemes()
		changed = true
	}
	if cfg.Theme == "" {
		cfg.Theme = DefaultThemeName
		changed = true
	}
	if _, ok := cfg.Themes[cfg.Theme]; !ok {
		cfg.Theme = DefaultThemeName
		if _, ok := cfg.Themes[cfg.Theme]; !ok {
			cfg.Theme = cfg.ThemeNames()[0]
		}
		changed = true
	}
	if cfg.CLI.Accent == "" {
		cfg.CLI.Accent = defaultCLIColors.Accent
		changed = true
	}
	if cfg.CLI.Error == "" {
		cfg.CLI.Error = defaultCLIColors.Error
		changed = true
	}
	if cfg.CLI.Success == "" {
		cfg.CLI.Success = defaultCLIColors.Success
		changed = true
	}
	if cfg.CLI.Warning == "" {
		cfg.CLI.Warning = defaultCLIColors.Warning
		changed = true
	}

	return changed
}

// SaveConfig writes the configuration to the default config path.
func SaveConfig(cfg *Config) error {
	return SaveConfigTo(ConfigPath(), cfg)
}

// SaveConfigTo writes cfg to path, preserving any unknown top-level keys that
// are already present in the file.
func SaveConfigTo(path string, cfg *Config) error {
	if path == "" {
		return fmt.Errorf("no config path available")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	// Preserve unknown top-level keys from the existing file.
	out := map[string]json.RawMessage{}
	if data, err := os.ReadFile(path); err == nil {
		var existing map[string]json.RawMessage
		if json.Unmarshal(data, &existing) == nil {
			out = existing
		}
	}

	themeRaw, err := json.Marshal(cfg.Theme)
	if err != nil {
		return err
	}
	themesRaw, err := json.Marshal(cfg.Themes)
	if err != nil {
		return err
	}
	cliRaw, err := json.Marshal(cfg.CLI)
	if err != nil {
		return err
	}
	out["theme"] = themeRaw
	out["themes"] = themesRaw
	out["cli"] = cliRaw

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0600)
}

// ThemeNames returns the names of every configured theme, sorted.
func (c *Config) ThemeNames() []string {
	names := make([]string, 0, len(c.Themes))
	for name := range c.Themes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Resolve returns the theme with the given name. If the name is unknown it
// falls back to the default theme, then to the first available theme.
func (c *Config) Resolve(name string) Theme {
	if t, ok := c.Themes[name]; ok {
		return t
	}
	if t, ok := c.Themes[DefaultThemeName]; ok {
		return t
	}
	if t, ok := c.Themes[DefaultConfig().Theme]; ok {
		return t
	}
	names := c.ThemeNames()
	if len(names) > 0 {
		return c.Themes[names[0]]
	}
	return Theme{}
}

// Active returns the currently selected theme.
func (c *Config) Active() Theme {
	return c.Resolve(c.Theme)
}

// EffectiveCLI returns the CLI colors with defaults applied for any blank field.
func (c *Config) EffectiveCLI() CLIColors {
	out := c.CLI
	if out.Accent == "" {
		out.Accent = defaultCLIColors.Accent
	}
	if out.Error == "" {
		out.Error = defaultCLIColors.Error
	}
	if out.Success == "" {
		out.Success = defaultCLIColors.Success
	}
	if out.Warning == "" {
		out.Warning = defaultCLIColors.Warning
	}
	return out
}
