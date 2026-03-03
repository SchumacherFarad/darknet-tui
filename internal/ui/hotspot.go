package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/SchumacherFarad/darknet-tui/internal/nmcli"
	"github.com/SchumacherFarad/darknet-tui/pkg/theme"
)

// HotspotView provides hotspot configuration and management.
type HotspotView struct {
	*tview.Flex
	theme   *theme.Theme
	adapter *nmcli.Adapter
	app     *tview.Application

	// Components
	statusPanel *tview.TextView
	configForm  *tview.Form

	// State
	device   *nmcli.Device
	ssid     string
	password string
	isActive bool
}

// NewHotspotView creates a new hotspot view.
func NewHotspotView(app *tview.Application, adapter *nmcli.Adapter, t *theme.Theme) *HotspotView {
	h := &HotspotView{
		Flex:     tview.NewFlex(),
		theme:    t,
		adapter:  adapter,
		app:      app,
		ssid:     "DarkNet-Hotspot",
		password: "",
	}

	h.setupUI()
	return h
}

func (h *HotspotView) setupUI() {
	// Status panel
	h.statusPanel = tview.NewTextView()
	h.statusPanel.SetDynamicColors(true)
	h.statusPanel.SetBackgroundColor(h.theme.Background)
	h.statusPanel.SetTextColor(h.theme.Foreground)
	h.statusPanel.SetBorder(true)
	h.statusPanel.SetTitle(" Hotspot Status ")
	h.statusPanel.SetTitleColor(h.theme.Title)
	h.statusPanel.SetBorderColor(h.theme.Border)

	// Configuration form
	h.configForm = tview.NewForm()
	h.configForm.SetBackgroundColor(h.theme.Background)
	h.configForm.SetFieldBackgroundColor(h.theme.StatusBar)
	h.configForm.SetFieldTextColor(h.theme.Foreground)
	h.configForm.SetButtonBackgroundColor(h.theme.Primary)
	h.configForm.SetButtonTextColor(h.theme.Background)
	h.configForm.SetLabelColor(h.theme.Foreground)
	h.configForm.SetBorder(true)
	h.configForm.SetTitle(" Configuration ")
	h.configForm.SetTitleColor(h.theme.Title)
	h.configForm.SetBorderColor(h.theme.Border)

	h.configForm.AddInputField("SSID:", h.ssid, 30, nil, func(text string) {
		h.ssid = text
	})

	h.configForm.AddPasswordField("Password:", h.password, 30, '*', func(text string) {
		h.password = text
	})

	h.configForm.AddButton("Start Hotspot", func() {
		h.startHotspot()
	})

	h.configForm.AddButton("Stop Hotspot", func() {
		h.stopHotspot()
	})

	// Layout
	h.Flex.SetDirection(tview.FlexRow)
	h.Flex.AddItem(h.statusPanel, 10, 0, false)
	h.Flex.AddItem(h.configForm, 0, 1, true)
}

// SetDevice sets the WiFi device for hotspot.
func (h *HotspotView) SetDevice(device *nmcli.Device) {
	h.device = device
	h.Refresh()
}

// Refresh updates the hotspot status.
func (h *HotspotView) Refresh() {
	if h.device == nil {
		h.statusPanel.SetText("[red]No WiFi device selected[-]")
		return
	}

	// Check for active hotspot connections
	activeConns, _ := h.adapter.GetAllActiveConnections()
	h.isActive = false

	for _, conn := range activeConns {
		if conn.Type == "802-11-wireless" {
			for _, devPath := range conn.Devices {
				if devPath == h.device.Path {
					// Check if this is a hotspot (AP mode)
					// For simplicity, check if connection name contains "Hotspot"
					if len(conn.ID) > 0 && containsHotspot(conn.ID) {
						h.isActive = true
						break
					}
				}
			}
		}
	}

	var statusText string
	if h.isActive {
		statusText = fmt.Sprintf(
			"[%s]Status:[-] [green]Active[-]\n\n"+
				"[%s]SSID:[-] %s\n\n"+
				"[%s]Device:[-] %s\n\n"+
				"[%s]Clients can connect to share your network[-]",
			colorToTag(h.theme.Primary),
			colorToTag(h.theme.Primary), h.ssid,
			colorToTag(h.theme.Primary), h.device.Interface,
			colorToTag(h.theme.Disconnected),
		)
	} else {
		statusText = fmt.Sprintf(
			"[%s]Status:[-] [gray]Inactive[-]\n\n"+
				"[%s]Device:[-] %s\n\n"+
				"[%s]Configure and start a hotspot to share your network[-]",
			colorToTag(h.theme.Primary),
			colorToTag(h.theme.Primary), h.device.Interface,
			colorToTag(h.theme.Disconnected),
		)
	}

	h.statusPanel.SetText(statusText)
}

func (h *HotspotView) startHotspot() {
	if h.device == nil {
		return
	}

	if h.ssid == "" {
		h.showError("SSID cannot be empty")
		return
	}

	if len(h.password) > 0 && len(h.password) < 8 {
		h.showError("Password must be at least 8 characters")
		return
	}

	err := h.adapter.CreateHotspot(h.device.Path, h.ssid, h.password)
	if err != nil {
		h.showError(fmt.Sprintf("Failed to create hotspot: %v", err))
		return
	}

	h.isActive = true
	h.Refresh()
}

func (h *HotspotView) stopHotspot() {
	if h.device == nil {
		return
	}

	// Find and deactivate hotspot connection
	activeConns, _ := h.adapter.GetAllActiveConnections()

	for _, conn := range activeConns {
		if conn.Type == "802-11-wireless" {
			for _, devPath := range conn.Devices {
				if devPath == h.device.Path && containsHotspot(conn.ID) {
					_ = h.adapter.DeactivateConnection(conn.Path)
					break
				}
			}
		}
	}

	h.isActive = false
	h.Refresh()
}

func (h *HotspotView) showError(message string) {
	// Create error modal
	modal := tview.NewModal()
	modal.SetText(message)
	modal.AddButtons([]string{"OK"})
	modal.SetBackgroundColor(h.theme.Background)
	modal.SetTextColor(h.theme.Error)
	modal.SetButtonBackgroundColor(h.theme.Primary)
	modal.SetButtonTextColor(h.theme.Background)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		h.app.SetRoot(h.getRootView(), true)
	})

	h.app.SetRoot(modal, true)
}

func (h *HotspotView) getRootView() tview.Primitive {
	return h.Flex
}

// HandleInput handles keyboard input for the hotspot view.
func (h *HotspotView) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'r', 'R':
			h.Refresh()
			return nil
		}
	}

	return event
}

// containsHotspot checks if a string contains "Hotspot" (case insensitive).
func containsHotspot(s string) bool {
	lower := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		lower[i] = c
	}

	target := "hotspot"
	if len(lower) < len(target) {
		return false
	}

	for i := 0; i <= len(lower)-len(target); i++ {
		match := true
		for j := 0; j < len(target); j++ {
			if lower[i+j] != target[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
