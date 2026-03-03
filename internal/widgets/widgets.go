// Package widgets provides custom tview widgets for darknet-tui.
package widgets

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/SchumacherFarad/darknet-tui/pkg/theme"
)

// SignalGauge is a custom widget that displays signal strength.
type SignalGauge struct {
	*tview.Box
	strength int
	theme    *theme.Theme
	label    string
}

// NewSignalGauge creates a new signal gauge widget.
func NewSignalGauge(t *theme.Theme) *SignalGauge {
	return &SignalGauge{
		Box:      tview.NewBox(),
		strength: 0,
		theme:    t,
	}
}

// SetStrength sets the signal strength (0-100).
func (s *SignalGauge) SetStrength(strength int) *SignalGauge {
	if strength < 0 {
		strength = 0
	}
	if strength > 100 {
		strength = 100
	}
	s.strength = strength
	return s
}

// SetLabel sets the label for the gauge.
func (s *SignalGauge) SetLabel(label string) *SignalGauge {
	s.label = label
	return s
}

// Draw renders the signal gauge.
func (s *SignalGauge) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen, s)
	x, y, width, _ := s.GetInnerRect()

	// Calculate bar representation
	bars := []rune{'▁', '▂', '▃', '▅', '▇'}
	numBars := (s.strength + 19) / 20
	if numBars > 5 {
		numBars = 5
	}

	// Get color based on strength
	color := s.theme.SignalColor(s.strength)
	style := tcell.StyleDefault.Foreground(color)
	dimStyle := tcell.StyleDefault.Foreground(s.theme.Disconnected)

	// Draw label if set
	labelOffset := 0
	if s.label != "" {
		labelStyle := tcell.StyleDefault.Foreground(s.theme.Foreground)
		for i, r := range s.label {
			if x+i < x+width-7 {
				screen.SetContent(x+i, y, r, nil, labelStyle)
			}
		}
		labelOffset = len(s.label) + 1
	}

	// Draw bars
	for i := 0; i < 5; i++ {
		posX := x + labelOffset + i
		if posX >= x+width {
			break
		}
		if i < numBars {
			screen.SetContent(posX, y, bars[i], nil, style)
		} else {
			screen.SetContent(posX, y, bars[i], nil, dimStyle)
		}
	}

	// Draw percentage
	pctStr := fmt.Sprintf(" %3d%%", s.strength)
	pctStyle := tcell.StyleDefault.Foreground(s.theme.Foreground)
	for i, r := range pctStr {
		posX := x + labelOffset + 5 + i
		if posX < x+width {
			screen.SetContent(posX, y, r, nil, pctStyle)
		}
	}
}

// StatusIndicator is a widget that shows connection status.
type StatusIndicator struct {
	*tview.Box
	connected  bool
	connecting bool
	theme      *theme.Theme
	label      string
}

// NewStatusIndicator creates a new status indicator widget.
func NewStatusIndicator(t *theme.Theme) *StatusIndicator {
	return &StatusIndicator{
		Box:   tview.NewBox(),
		theme: t,
	}
}

// SetConnected sets the connected state.
func (s *StatusIndicator) SetConnected(connected bool) *StatusIndicator {
	s.connected = connected
	s.connecting = false
	return s
}

// SetConnecting sets the connecting state.
func (s *StatusIndicator) SetConnecting(connecting bool) *StatusIndicator {
	s.connecting = connecting
	if connecting {
		s.connected = false
	}
	return s
}

// SetLabel sets the label for the indicator.
func (s *StatusIndicator) SetLabel(label string) *StatusIndicator {
	s.label = label
	return s
}

// Draw renders the status indicator.
func (s *StatusIndicator) Draw(screen tcell.Screen) {
	s.Box.DrawForSubclass(screen, s)
	x, y, width, _ := s.GetInnerRect()

	var dot rune
	var color tcell.Color

	switch {
	case s.connected:
		dot = '●'
		color = s.theme.Connected
	case s.connecting:
		dot = '◐'
		color = s.theme.Connecting
	default:
		dot = '○'
		color = s.theme.Disconnected
	}

	// Draw dot
	dotStyle := tcell.StyleDefault.Foreground(color)
	screen.SetContent(x, y, dot, nil, dotStyle)

	// Draw label
	if s.label != "" {
		labelStyle := tcell.StyleDefault.Foreground(s.theme.Foreground)
		for i, r := range s.label {
			if x+2+i < x+width {
				screen.SetContent(x+2+i, y, r, nil, labelStyle)
			}
		}
	}
}

// NetworkList is a styled list for network items.
type NetworkList struct {
	*tview.List
	theme *theme.Theme
}

// NewNetworkList creates a new network list widget.
func NewNetworkList(t *theme.Theme) *NetworkList {
	list := tview.NewList()
	list.ShowSecondaryText(true)
	list.SetHighlightFullLine(true)
	list.SetBorderPadding(0, 0, 1, 1)

	// Apply theme colors
	list.SetBackgroundColor(t.Background)
	list.SetMainTextColor(t.Foreground)
	list.SetSecondaryTextColor(t.Disconnected)
	list.SetSelectedBackgroundColor(t.Selection)
	list.SetSelectedTextColor(t.Foreground)

	return &NetworkList{
		List:  list,
		theme: t,
	}
}

// AddNetwork adds a network item to the list.
func (n *NetworkList) AddNetwork(name, details string, connected bool, handler func()) *NetworkList {
	var prefix string
	if connected {
		prefix = "[green]●[white] "
	} else {
		prefix = "[gray]○[white] "
	}

	n.List.AddItem(prefix+name, details, 0, handler)
	return n
}

// CommandPalette is a search/command input widget.
type CommandPalette struct {
	*tview.InputField
	theme    *theme.Theme
	commands []Command
	onSelect func(Command)
}

// Command represents a command palette item.
type Command struct {
	Name        string
	Description string
	Action      func()
}

// NewCommandPalette creates a new command palette widget.
func NewCommandPalette(t *theme.Theme) *CommandPalette {
	input := tview.NewInputField()
	input.SetLabel("  ")
	input.SetFieldBackgroundColor(t.StatusBar)
	input.SetFieldTextColor(t.Foreground)
	input.SetLabelColor(t.Primary)
	input.SetPlaceholder("Type a command...")
	input.SetPlaceholderTextColor(t.Disconnected)

	return &CommandPalette{
		InputField: input,
		theme:      t,
		commands:   make([]Command, 0),
	}
}

// SetCommands sets the available commands.
func (c *CommandPalette) SetCommands(commands []Command) *CommandPalette {
	c.commands = commands
	return c
}

// SetOnSelect sets the handler for command selection.
func (c *CommandPalette) SetOnSelect(handler func(Command)) *CommandPalette {
	c.onSelect = handler
	return c
}

// FilterCommands returns commands matching the query.
func (c *CommandPalette) FilterCommands(query string) []Command {
	if query == "" {
		return c.commands
	}

	query = strings.ToLower(query)
	filtered := make([]Command, 0)
	for _, cmd := range c.commands {
		if strings.Contains(strings.ToLower(cmd.Name), query) ||
			strings.Contains(strings.ToLower(cmd.Description), query) {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

// StatusBar is a bottom status bar widget.
type StatusBar struct {
	*tview.TextView
	theme      *theme.Theme
	leftText   string
	centerText string
	rightText  string
}

// NewStatusBar creates a new status bar widget.
func NewStatusBar(t *theme.Theme) *StatusBar {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBackgroundColor(t.StatusBar)
	tv.SetTextColor(t.Foreground)

	return &StatusBar{
		TextView: tv,
		theme:    t,
	}
}

// SetLeft sets the left text.
func (s *StatusBar) SetLeft(text string) *StatusBar {
	s.leftText = text
	s.updateText()
	return s
}

// SetCenter sets the center text.
func (s *StatusBar) SetCenter(text string) *StatusBar {
	s.centerText = text
	s.updateText()
	return s
}

// SetRight sets the right text.
func (s *StatusBar) SetRight(text string) *StatusBar {
	s.rightText = text
	s.updateText()
	return s
}

func (s *StatusBar) updateText() {
	_, _, width, _ := s.GetInnerRect()
	if width <= 0 {
		width = 80 // default
	}

	// Calculate padding
	leftLen := len(s.leftText)
	centerLen := len(s.centerText)
	rightLen := len(s.rightText)

	centerPad := (width - centerLen) / 2
	if centerPad < leftLen+1 {
		centerPad = leftLen + 1
	}

	rightPad := width - rightLen
	if rightPad < centerPad+centerLen+1 {
		rightPad = centerPad + centerLen + 1
	}

	// Build status line
	line := s.leftText
	line += strings.Repeat(" ", centerPad-leftLen)
	line += s.centerText
	line += strings.Repeat(" ", rightPad-centerPad-centerLen)
	line += s.rightText

	s.TextView.SetText(line)
}
