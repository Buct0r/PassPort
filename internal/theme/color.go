// Package theme provides the shared, Fyne-free theme and configuration model
// used by both the GUI (src) and CLI (cli) binaries of PassPort.
package theme

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// ansiPalette holds the RGB approximations of the standard 16 ANSI colors,
// used to pick the closest color when truecolor is unavailable.
var ansiPalette = [16]struct{ r, g, b uint8 }{
	{0, 0, 0},       // 0  black
	{128, 0, 0},     // 1  red
	{0, 128, 0},     // 2  green
	{128, 128, 0},   // 3  yellow
	{0, 0, 128},     // 4  blue
	{128, 0, 128},   // 5  magenta
	{0, 128, 128},   // 6  cyan
	{192, 192, 192}, // 7  white
	{128, 128, 128}, // 8  bright black
	{255, 0, 0},     // 9  bright red
	{0, 255, 0},     // 10 bright green
	{255, 255, 0},   // 11 bright yellow
	{0, 0, 255},     // 12 bright blue
	{255, 0, 255},   // 13 bright magenta
	{0, 255, 255},   // 14 bright cyan
	{255, 255, 255}, // 15 bright white
}

func hexByte(s string) (uint8, bool) {
	v, err := strconv.ParseUint(s, 16, 8)
	if err != nil {
		return 0, false
	}
	return uint8(v), true
}

// ParseHex parses a hex color string into a color.RGBA. It accepts "#RGB",
// "#RRGGBB" and "#RRGGBBAA", with or without the leading '#'. The boolean
// result reports whether the string was a valid color.
func ParseHex(s string) (color.RGBA, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")

	switch len(s) {
	case 3:
		r, ok1 := hexByte(string([]byte{s[0], s[0]}))
		g, ok2 := hexByte(string([]byte{s[1], s[1]}))
		b, ok3 := hexByte(string([]byte{s[2], s[2]}))
		if !ok1 || !ok2 || !ok3 {
			return color.RGBA{}, false
		}
		return color.RGBA{R: r, G: g, B: b, A: 255}, true
	case 6:
		r, ok1 := hexByte(s[0:2])
		g, ok2 := hexByte(s[2:4])
		b, ok3 := hexByte(s[4:6])
		if !ok1 || !ok2 || !ok3 {
			return color.RGBA{}, false
		}
		return color.RGBA{R: r, G: g, B: b, A: 255}, true
	case 8:
		r, ok1 := hexByte(s[0:2])
		g, ok2 := hexByte(s[2:4])
		b, ok3 := hexByte(s[4:6])
		a, ok4 := hexByte(s[6:8])
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return color.RGBA{}, false
		}
		return color.RGBA{R: r, G: g, B: b, A: a}, true
	default:
		return color.RGBA{}, false
	}
}

// SupportsTrueColor reports whether the current terminal is expected to
// understand 24-bit ("truecolor") ANSI escape sequences. It also honours the
// NO_COLOR convention, in which case no escape sequences should be emitted.
func SupportsTrueColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	switch os.Getenv("COLORTERM") {
	case "truecolor", "24bit":
		return true
	}
	term := os.Getenv("TERM")
	if strings.Contains(term, "truecolor") || strings.Contains(term, "24bit") || strings.Contains(term, "direct") {
		return true
	}
	// Windows Terminal and VS Code's integrated terminal both support it.
	if os.Getenv("WT_SESSION") != "" || os.Getenv("TERM_PROGRAM") == "vscode" {
		return true
	}
	return false
}

// ColorEnabled reports whether colored output should be produced at all.
func ColorEnabled() bool {
	return os.Getenv("NO_COLOR") == ""
}

func nearestANSI(c color.RGBA) int {
	best := 0
	bestDist := 1 << 30
	for i, p := range ansiPalette {
		dr := int(c.R) - int(p.r)
		dg := int(c.G) - int(p.g)
		db := int(c.B) - int(p.b)
		if d := dr*dr + dg*dg + db*db; d < bestDist {
			bestDist = d
			best = i
		}
	}
	if best < 8 {
		return 30 + best
	}
	return 90 + (best - 8)
}

// ANSIColor converts a hex color into an ANSI escape sequence for the
// foreground. When truecolor is unsupported it falls back to the nearest of
// the 16 basic colors; when the hex value is invalid it uses fallback (a raw
// SGR code such as 36 for cyan). When colors are disabled it returns "".
func ANSIColor(hex string, fallback int) string {
	if !ColorEnabled() {
		return ""
	}
	c, ok := ParseHex(hex)
	if !ok {
		return fmt.Sprintf("\033[%dm", fallback)
	}
	if SupportsTrueColor() {
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", c.R, c.G, c.B)
	}
	return fmt.Sprintf("\033[%dm", nearestANSI(c))
}

// Reset is the ANSI escape sequence that restores default terminal colors.
func Reset() string {
	if !ColorEnabled() {
		return ""
	}
	return "\033[0m"
}
