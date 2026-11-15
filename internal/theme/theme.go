package theme

import "image/color"

// CardBackgroundKey is the custom theme key for the password card background.
// It is not part of Fyne's own color set, so it is handled separately by the
// GUI adapter.
const CardBackgroundKey = "cardBackground"

// Theme maps color keys to hex color strings. Keys correspond to Fyne's
// ThemeColorName string values (e.g. "background", "primary", "foreground")
// plus the custom CardBackgroundKey. Themes may be partial; any key that is
// absent falls back to Fyne's default theme in the GUI adapter.
type Theme map[string]string

// Get returns the raw hex value stored for key.
func (t Theme) Get(key string) (string, bool) {
	if t == nil {
		return "", false
	}
	v, ok := t[key]
	return v, ok
}

// Color resolves key to a color.RGBA, reporting whether it was present and valid.
func (t Theme) Color(key string) (color.RGBA, bool) {
	v, ok := t.Get(key)
	if !ok {
		return color.RGBA{}, false
	}
	return ParseHex(v)
}

// Clone returns a deep copy of the theme.
func (t Theme) Clone() Theme {
	out := make(Theme, len(t))
	for k, v := range t {
		out[k] = v
	}
	return out
}
