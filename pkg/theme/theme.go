// Package theme provides the Dark Blue color scheme for darknet-tui.
package theme

import "github.com/gdamore/tcell/v2"

// Theme contains all color definitions for the TUI.
type Theme struct {
	// Primary colors
	Background tcell.Color
	Foreground tcell.Color
	Primary    tcell.Color
	Secondary  tcell.Color
	Accent     tcell.Color

	// Status colors
	Connected    tcell.Color
	Connecting   tcell.Color
	Disconnected tcell.Color
	Error        tcell.Color

	// UI element colors
	Border      tcell.Color
	BorderFocus tcell.Color
	Title       tcell.Color
	StatusBar   tcell.Color
	Selection   tcell.Color

	// Signal strength colors
	SignalHigh   tcell.Color
	SignalMedium tcell.Color
	SignalLow    tcell.Color
	SignalWeak   tcell.Color
}

// DarkBlue returns the default Dark Blue theme.
func DarkBlue() *Theme {
	return &Theme{
		// Primary colors - Dark blue palette
		Background: tcell.NewHexColor(0x0d1117),
		Foreground: tcell.NewHexColor(0xc9d1d9),
		Primary:    tcell.NewHexColor(0x58a6ff),
		Secondary:  tcell.NewHexColor(0x388bfd),
		Accent:     tcell.NewHexColor(0x79c0ff),

		// Status colors
		Connected:    tcell.NewHexColor(0x3fb950),
		Connecting:   tcell.NewHexColor(0xd29922),
		Disconnected: tcell.NewHexColor(0x8b949e),
		Error:        tcell.NewHexColor(0xf85149),

		// UI element colors
		Border:      tcell.NewHexColor(0x30363d),
		BorderFocus: tcell.NewHexColor(0x58a6ff),
		Title:       tcell.NewHexColor(0x58a6ff),
		StatusBar:   tcell.NewHexColor(0x161b22),
		Selection:   tcell.NewHexColor(0x1f6feb),

		// Signal strength colors
		SignalHigh:   tcell.NewHexColor(0x3fb950),
		SignalMedium: tcell.NewHexColor(0x58a6ff),
		SignalLow:    tcell.NewHexColor(0xd29922),
		SignalWeak:   tcell.NewHexColor(0xf85149),
	}
}

// SignalGauge returns a visual representation of signal strength.
// Strength should be 0-100 (percentage).
func SignalGauge(strength int) string {
	bars := []rune{'▁', '▂', '▃', '▅', '▇'}

	if strength <= 0 {
		return "     "
	}

	// Calculate how many bars to show (1-5)
	numBars := (strength + 19) / 20 // Maps 1-20 to 1, 21-40 to 2, etc.
	if numBars > 5 {
		numBars = 5
	}

	result := ""
	for i := 0; i < 5; i++ {
		if i < numBars {
			result += string(bars[i])
		} else {
			result += " "
		}
	}
	return result
}

// SignalColor returns the appropriate color for signal strength.
func (t *Theme) SignalColor(strength int) tcell.Color {
	switch {
	case strength >= 70:
		return t.SignalHigh
	case strength >= 50:
		return t.SignalMedium
	case strength >= 30:
		return t.SignalLow
	default:
		return t.SignalWeak
	}
}

// ConnectionDot returns the appropriate status indicator.
func ConnectionDot(connected bool) string {
	if connected {
		return "●"
	}
	return "○"
}

// NetworkIcon returns an icon for the network type.
func NetworkIcon(networkType string) string {
	switch networkType {
	case "wifi":
		return "󰖩"
	case "ethernet":
		return "󰈀"
	case "vpn":
		return "󰒄"
	case "hotspot":
		return "󱛁"
	default:
		return "󰛳"
	}
}
