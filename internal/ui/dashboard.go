// Package ui provides the user interface components for darknet-tui.
package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/SchumacherFarad/darknet-tui/internal/nmcli"
	"github.com/SchumacherFarad/darknet-tui/internal/widgets"
	"github.com/SchumacherFarad/darknet-tui/pkg/theme"
)

// Dashboard is the main dashboard view.
type Dashboard struct {
	*tview.Flex
	theme   *theme.Theme
	adapter *nmcli.Adapter
	app     *tview.Application

	// Panels
	statusPanel  *tview.TextView
	devicePanel  *tview.List
	networkPanel *tview.Flex
	statusBar    *widgets.StatusBar

	// Current view
	wifiView     *WifiView
	ethernetView *EthernetView
	hotspotView  *HotspotView

	// State
	currentPanel       int
	panels             []tview.Primitive
	selectedDevicePath string
}

// NewDashboard creates a new dashboard view.
func NewDashboard(app *tview.Application, adapter *nmcli.Adapter, t *theme.Theme) *Dashboard {
	d := &Dashboard{
		Flex:    tview.NewFlex(),
		theme:   t,
		adapter: adapter,
		app:     app,
	}

	d.setupUI()
	return d
}

func (d *Dashboard) setupUI() {
	// Status panel (top)
	d.statusPanel = tview.NewTextView()
	d.statusPanel.SetDynamicColors(true)
	d.statusPanel.SetBorder(true)
	d.statusPanel.SetTitle(" Network Status ")
	d.statusPanel.SetTitleColor(d.theme.Title)
	d.statusPanel.SetBorderColor(d.theme.Border)
	d.statusPanel.SetBackgroundColor(d.theme.Background)
	d.statusPanel.SetTextColor(d.theme.Foreground)

	// Device list panel (left sidebar)
	d.devicePanel = tview.NewList()
	d.devicePanel.ShowSecondaryText(true)
	d.devicePanel.SetHighlightFullLine(true)
	d.devicePanel.SetBorder(true)
	d.devicePanel.SetTitle(" Devices ")
	d.devicePanel.SetTitleColor(d.theme.Title)
	d.devicePanel.SetBorderColor(d.theme.Border)
	d.devicePanel.SetBackgroundColor(d.theme.Background)
	d.devicePanel.SetMainTextColor(d.theme.Foreground)
	d.devicePanel.SetSecondaryTextColor(d.theme.Disconnected)
	d.devicePanel.SetSelectedBackgroundColor(d.theme.Selection)
	d.devicePanel.SetSelectedTextColor(d.theme.Foreground)

	// Network panel (main content area)
	d.networkPanel = tview.NewFlex()
	d.networkPanel.SetBorder(true)
	d.networkPanel.SetTitle(" WiFi Networks ")
	d.networkPanel.SetTitleColor(d.theme.Title)
	d.networkPanel.SetBorderColor(d.theme.Border)
	d.networkPanel.SetBackgroundColor(d.theme.Background)

	// Initialize views
	d.wifiView = NewWifiView(d.app, d.adapter, d.theme)
	d.ethernetView = NewEthernetView(d.app, d.adapter, d.theme)
	d.hotspotView = NewHotspotView(d.app, d.adapter, d.theme)

	// Set WiFi as default view
	d.networkPanel.AddItem(d.wifiView, 0, 1, true)

	// Status bar (bottom)
	d.statusBar = widgets.NewStatusBar(d.theme)
	d.statusBar.SetLeft(" darknet-tui")
	d.statusBar.SetCenter("Tab: Switch Panel | /: Search | q: Quit")
	d.statusBar.SetRight("Ctrl+P: Commands ")

	// Build layout
	// Top row: Status panel
	// Middle row: Device list (left) + Network panel (right)
	// Bottom row: Status bar

	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	leftPanel.AddItem(d.devicePanel, 0, 1, false)

	mainContent := tview.NewFlex()
	mainContent.AddItem(leftPanel, 25, 0, false)
	mainContent.AddItem(d.networkPanel, 0, 1, true)

	d.Flex.SetDirection(tview.FlexRow)
	d.Flex.AddItem(d.statusPanel, 5, 0, false)
	d.Flex.AddItem(mainContent, 0, 1, true)
	d.Flex.AddItem(d.statusBar, 1, 0, false)

	// Track focusable panels
	d.panels = []tview.Primitive{d.devicePanel, d.wifiView}
	d.currentPanel = 1 // Start with network panel focused

	// Setup device list selection handler
	d.devicePanel.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		d.onDeviceSelected(index)
	})

	// Load initial data
	d.Refresh()
}

// Refresh reloads all data from NetworkManager.
func (d *Dashboard) Refresh() {
	d.refreshStatus()
	d.refreshDevices()
	d.wifiView.Refresh()
	d.ethernetView.Refresh()
}

func (d *Dashboard) refreshStatus() {
	state, _ := d.adapter.GetState()
	version, _ := d.adapter.GetVersion()
	wirelessEnabled, _ := d.adapter.WirelessEnabled()

	wirelessStatus := "[red]Disabled"
	if wirelessEnabled {
		wirelessStatus = "[green]Enabled"
	}

	statusText := fmt.Sprintf(
		"[%s]NetworkManager[-] v%s\n"+
			"[%s]State:[-] %s\n"+
			"[%s]Wireless:[-] %s",
		colorToTag(d.theme.Primary), version,
		colorToTag(d.theme.Primary), state.String(),
		colorToTag(d.theme.Primary), wirelessStatus,
	)

	d.statusPanel.SetText(statusText)
}

func (d *Dashboard) refreshDevices() {
	// Save current selection
	currentIndex := d.devicePanel.GetCurrentItem()
	var currentInterface string
	devices, err := d.adapter.GetAllDevices()
	if err == nil && currentIndex >= 0 && currentIndex < len(devices) {
		currentInterface = devices[currentIndex].Interface
	}

	d.devicePanel.Clear()

	if err != nil {
		return
	}

	for _, device := range devices {
		var icon string
		switch device.Type {
		case nmcli.DeviceTypeWifi:
			icon = "󰖩"
		case nmcli.DeviceTypeEthernet:
			icon = "󰈀"
		default:
			icon = "󰛳"
		}

		status := device.State.String()
		mainText := fmt.Sprintf("%s %s", icon, device.Interface)
		secondaryText := fmt.Sprintf("  %s - %s", device.Type.String(), status)

		d.devicePanel.AddItem(mainText, secondaryText, 0, nil)
	}

	// Restore selection
	targetIndex := 0
	for i, device := range devices {
		if device.Interface == currentInterface {
			targetIndex = i
			break
		}
	}
	d.devicePanel.SetCurrentItem(targetIndex)
}

func (d *Dashboard) onDeviceSelected(index int) {
	devices, err := d.adapter.GetAllDevices()
	if err != nil || index >= len(devices) {
		return
	}

	device := devices[index]

	// Clear and update network panel based on device type
	d.networkPanel.Clear()

	switch device.Type {
	case nmcli.DeviceTypeWifi:
		d.networkPanel.SetTitle(" WiFi Networks ")
		d.wifiView.SetDevice(device)
		d.networkPanel.AddItem(d.wifiView, 0, 1, true)
		d.panels[1] = d.wifiView
	case nmcli.DeviceTypeEthernet:
		d.networkPanel.SetTitle(" Ethernet ")
		d.ethernetView.SetDevice(device)
		d.networkPanel.AddItem(d.ethernetView, 0, 1, true)
		d.panels[1] = d.ethernetView
	default:
		d.networkPanel.SetTitle(" Device Details ")
	}

	d.app.SetFocus(d.panels[1])
}

// HandleInput handles keyboard input for the dashboard.
func (d *Dashboard) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		// Cycle through panels
		d.currentPanel = (d.currentPanel + 1) % len(d.panels)
		d.app.SetFocus(d.panels[d.currentPanel])

		// Update border colors to show focus
		if d.currentPanel == 0 {
			d.devicePanel.SetBorderColor(d.theme.BorderFocus)
			d.networkPanel.SetBorderColor(d.theme.Border)
		} else {
			d.devicePanel.SetBorderColor(d.theme.Border)
			d.networkPanel.SetBorderColor(d.theme.BorderFocus)
		}
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			d.app.Stop()
			return nil
		case 'r', 'R':
			d.Refresh()
			return nil
		}
	}

	return event
}

// GetFocusable returns the current focusable panel.
func (d *Dashboard) GetFocusable() tview.Primitive {
	return d.panels[d.currentPanel]
}

// colorToTag converts a tcell color to a tview color tag.
func colorToTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
