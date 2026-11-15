package theme

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseHex(t *testing.T) {
	cases := []struct {
		in      string
		ok      bool
		r, g, b uint8
		a       uint8
	}{
		{"#FFFFFF", true, 255, 255, 255, 255},
		{"FFFFFF", true, 255, 255, 255, 255},
		{"#000000", true, 0, 0, 0, 255},
		{"#F00", true, 255, 0, 0, 255},
		{"#00ff00", true, 0, 255, 0, 255},
		{"#01020304", true, 1, 2, 3, 4},
		{"  #0A1E32  ", true, 10, 30, 50, 255},
		{"", false, 0, 0, 0, 0},
		{"#12", false, 0, 0, 0, 0},
		{"#GGGGGG", false, 0, 0, 0, 0},
		{"#1234567", false, 0, 0, 0, 0},
	}

	for _, c := range cases {
		got, ok := ParseHex(c.in)
		if ok != c.ok {
			t.Errorf("ParseHex(%q) ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		if got.R != c.r || got.G != c.g || got.B != c.b || got.A != c.a {
			t.Errorf("ParseHex(%q) = %v, want %d,%d,%d,%d", c.in, got, c.r, c.g, c.b, c.a)
		}
	}
}

func TestANSIColorDisabled(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if got := ANSIColor("#FF0000", 31); got != "" {
		t.Errorf("ANSIColor with NO_COLOR = %q, want empty", got)
	}
	if got := Reset(); got != "" {
		t.Errorf("Reset with NO_COLOR = %q, want empty", got)
	}
}

func TestANSIColorFallback(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("COLORTERM", "")
	t.Setenv("TERM", "xterm")
	t.Setenv("WT_SESSION", "")
	t.Setenv("TERM_PROGRAM", "")

	if got := ANSIColor("not-a-color", 36); got != "\033[36m" {
		t.Errorf("ANSIColor invalid = %q, want \\033[36m", got)
	}

	got := ANSIColor("#FF0000", 31)
	if got == "" || got == "\033[31m" {
		t.Errorf("ANSIColor(#FF0000) = %q, expected nearest-color escape", got)
	}
}

func TestANSIColorTrueColor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	t.Setenv("COLORTERM", "truecolor")

	got := ANSIColor("#010203", 31)
	want := "\033[38;2;1;2;3m"
	if got != want {
		t.Errorf("ANSIColor = %q, want %q", got, want)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := DefaultConfig()
	cfg.Theme = "NordTheme"
	if err := SaveConfigTo(path, cfg); err != nil {
		t.Fatalf("SaveConfigTo: %v", err)
	}

	got, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if got.Theme != "NordTheme" {
		t.Errorf("Theme = %q, want NordTheme", got.Theme)
	}
	if len(got.Themes) != len(DefaultThemes()) {
		t.Errorf("themes count = %d, want %d", len(got.Themes), len(DefaultThemes()))
	}
	if got.CLI.Accent != defaultCLIColors.Accent {
		t.Errorf("CLI.Accent = %q, want %q", got.CLI.Accent, defaultCLIColors.Accent)
	}
}

func TestConfigMissingFileCreatesTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")

	cfg, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if cfg.Theme != DefaultThemeName {
		t.Errorf("Theme = %q, want %q", cfg.Theme, DefaultThemeName)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected config template to be written: %v", err)
	}
}

func TestConfigMigrationFromOldFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte(`{"theme":"LightTheme"}`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if cfg.Theme != "LightTheme" {
		t.Errorf("Theme = %q, want LightTheme", cfg.Theme)
	}
	if len(cfg.Themes) == 0 {
		t.Error("expected built-in themes to be injected during migration")
	}

	// The migration should have persisted the full theme set.
	raw, _ := os.ReadFile(path)
	var check map[string]json.RawMessage
	if err := json.Unmarshal(raw, &check); err != nil {
		t.Fatalf("migrated file is not valid JSON: %v", err)
	}
	if _, ok := check["themes"]; !ok {
		t.Error("expected 'themes' key to be written after migration")
	}
}

func TestConfigPreservesUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := os.WriteFile(path, []byte(`{"theme":"CustomTheme","futureFeature":42}`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if err := SaveConfigTo(path, cfg); err != nil {
		t.Fatalf("SaveConfigTo: %v", err)
	}

	raw, _ := os.ReadFile(path)
	var check map[string]json.RawMessage
	if err := json.Unmarshal(raw, &check); err != nil {
		t.Fatal(err)
	}
	if _, ok := check["futureFeature"]; !ok {
		t.Error("unknown key 'futureFeature' was not preserved")
	}
}

func TestConfigResolveFallback(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Theme = "DoesNotExist"

	if len(cfg.Resolve("DoesNotExist")) == 0 {
		t.Error("Resolve should fall back to a non-empty theme")
	}
	if len(cfg.Active()) == 0 {
		t.Error("Active should return a non-empty theme")
	}
}

func TestThemeNamesSorted(t *testing.T) {
	cfg := DefaultConfig()
	names := cfg.ThemeNames()
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("ThemeNames not sorted: %v", names)
			break
		}
	}
}
