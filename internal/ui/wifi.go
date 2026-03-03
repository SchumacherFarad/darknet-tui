package ui

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/SchumacherFarad/darknet-tui/internal/nmcli"
	"github.com/SchumacherFarad/darknet-tui/pkg/theme"
)

// WifiView displays WiFi networks and allows connection management.
type WifiView struct {
	*tview.Flex
	theme   *theme.Theme
	adapter *nmcli.Adapter
	app     *tview.Application

	// Components
	networkList  *tview.List
	detailsPanel *tview.TextView
	searchInput  *tview.InputField

	// State
	device       *nmcli.Device
	accessPoints []*nmcli.AccessPoint
	searching    bool
}

// NewWifiView creates a new WiFi view.
func NewWifiView(app *tview.Application, adapter *nmcli.Adapter, t *theme.Theme) *WifiView {
	w := &WifiView{
		Flex:    tview.NewFlex(),
		theme:   t,
		adapter: adapter,
		app:     app,
	}

	w.setupUI()
	return w
}

func (w *WifiView) setupUI() {
	// Network list
	w.networkList = tview.NewList()
	w.networkList.ShowSecondaryText(true)
	w.networkList.SetHighlightFullLine(true)
	w.networkList.SetBackgroundColor(w.theme.Background)
	w.networkList.SetMainTextColor(w.theme.Foreground)
	w.networkList.SetSecondaryTextColor(w.theme.Disconnected)
	w.networkList.SetSelectedBackgroundColor(w.theme.Selection)
	w.networkList.SetSelectedTextColor(w.theme.Foreground)
	w.networkList.SetBorderPadding(0, 0, 1, 1)

	// Details panel
	w.detailsPanel = tview.NewTextView()
	w.detailsPanel.SetDynamicColors(true)
	w.detailsPanel.SetBackgroundColor(w.theme.Background)
	w.detailsPanel.SetTextColor(w.theme.Foreground)
	w.detailsPanel.SetBorder(true)
	w.detailsPanel.SetTitle(" Details ")
	w.detailsPanel.SetTitleColor(w.theme.Title)
	w.detailsPanel.SetBorderColor(w.theme.Border)

	// Search input
	w.searchInput = tview.NewInputField()
	w.searchInput.SetLabel(" / ")
	w.searchInput.SetFieldBackgroundColor(w.theme.StatusBar)
	w.searchInput.SetFieldTextColor(w.theme.Foreground)
	w.searchInput.SetLabelColor(w.theme.Primary)
	w.searchInput.SetPlaceholder("Search networks...")
	w.searchInput.SetPlaceholderTextColor(w.theme.Disconnected)
	w.searchInput.SetBackgroundColor(w.theme.Background)

	// Layout
	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	leftPanel.AddItem(w.networkList, 0, 1, true)

	w.Flex.AddItem(leftPanel, 0, 2, true)
	w.Flex.AddItem(w.detailsPanel, 35, 0, false)

	// Selection handler
	w.networkList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		w.onNetworkSelected(index)
	})

	w.networkList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		w.onNetworkActivated(index)
	})
}

// SetDevice sets the WiFi device to use.
func (w *WifiView) SetDevice(device *nmcli.Device) {
	w.device = device
	w.Refresh()
}

// Refresh reloads the access point list.
func (w *WifiView) Refresh() {
	if w.device == nil {
		return
	}

	// Trigger scan
	_ = w.adapter.ScanWifi(w.device.Path)

	// Get access points
	aps, err := w.adapter.GetAccessPoints(w.device.Path)
	if err != nil {
		return
	}

	// Sort by signal strength (strongest first)
	sort.Slice(aps, func(i, j int) bool {
		return aps[i].Strength > aps[j].Strength
	})

	w.accessPoints = aps

	// Get active access point
	activeAP, _ := w.adapter.GetActiveAccessPoint(w.device.Path)

	// Update list
	w.networkList.Clear()

	for _, ap := range aps {
		if ap.SSID == "" {
			continue // Skip hidden networks
		}

		// Build display text
		signalGauge := theme.SignalGauge(int(ap.Strength))

		var statusIcon string
		if activeAP != nil && ap.BSSID == activeAP.BSSID {
			statusIcon = "[green]●[-] "
		} else {
			statusIcon = "[gray]○[-] "
		}

		securityIcon := ""
		if ap.Security > 0 {
			securityIcon = " 󰌾"
		}

		mainText := fmt.Sprintf("%s%s%s", statusIcon, ap.SSID, securityIcon)
		secondaryText := fmt.Sprintf("  %s %d%%  %dMHz", signalGauge, ap.Strength, ap.Freq)

		w.networkList.AddItem(mainText, secondaryText, 0, nil)
	}

	// Update details for first item
	if len(w.accessPoints) > 0 {
		w.onNetworkSelected(0)
	}
}

func (w *WifiView) onNetworkSelected(index int) {
	if index < 0 || index >= len(w.accessPoints) {
		return
	}

	ap := w.accessPoints[index]

	// Security type
	securityType := "Open"
	if ap.Security > 0 {
		securityType = "WPA/WPA2"
	}

	// Frequency band
	band := "2.4 GHz"
	if ap.Freq >= 5000 {
		band = "5 GHz"
	}

	details := fmt.Sprintf(
		"[%s]SSID:[-] %s\n\n"+
			"[%s]BSSID:[-] %s\n\n"+
			"[%s]Signal:[-] %d%%\n\n"+
			"[%s]Frequency:[-] %d MHz (%s)\n\n"+
			"[%s]Security:[-] %s\n\n"+
			"[%s]Press Enter to connect[-]",
		colorToTag(w.theme.Primary), ap.SSID,
		colorToTag(w.theme.Primary), ap.BSSID,
		colorToTag(w.theme.Primary), ap.Strength,
		colorToTag(w.theme.Primary), ap.Freq, band,
		colorToTag(w.theme.Primary), securityType,
		colorToTag(w.theme.Disconnected),
	)

	w.detailsPanel.SetText(details)
}

func (w *WifiView) onNetworkActivated(index int) {
	if index < 0 || index >= len(w.accessPoints) {
		return
	}

	ap := w.accessPoints[index]

	// Check if already connected
	activeAP, _ := w.adapter.GetActiveAccessPoint(w.device.Path)
	if activeAP != nil && ap.BSSID == activeAP.BSSID {
		// Disconnect
		conns, _ := w.adapter.GetAllActiveConnections()
		for _, conn := range conns {
			for _, devPath := range conn.Devices {
				if devPath == w.device.Path {
					_ = w.adapter.DeactivateConnection(conn.Path)
					w.Refresh()
					return
				}
			}
		}
		return
	}

	// If network requires password, show input dialog
	if ap.Security > 0 {
		w.showPasswordDialog(ap)
	} else {
		// Connect without password
		_ = w.adapter.ConnectToWifi(w.device.Path, ap.Path, ap.SSID, "")
		w.Refresh()
	}
}

func (w *WifiView) showPasswordDialog(ap *nmcli.AccessPoint) {
	// Create password input modal
	form := tview.NewForm()
	form.SetBackgroundColor(w.theme.Background)
	form.SetFieldBackgroundColor(w.theme.StatusBar)
	form.SetFieldTextColor(w.theme.Foreground)
	form.SetButtonBackgroundColor(w.theme.Primary)
	form.SetButtonTextColor(w.theme.Background)
	form.SetLabelColor(w.theme.Foreground)
	form.SetBorder(true)
	form.SetTitle(fmt.Sprintf(" Connect to %s ", ap.SSID))
	form.SetTitleColor(w.theme.Title)
	form.SetBorderColor(w.theme.BorderFocus)

	var password string

	form.AddPasswordField("Password:", "", 30, '*', func(text string) {
		password = text
	})

	form.AddButton("Connect", func() {
		_ = w.adapter.ConnectToWifi(w.device.Path, ap.Path, ap.SSID, password)
		w.app.SetRoot(w.getRootView(), true)
		w.Refresh()
	})

	form.AddButton("Cancel", func() {
		w.app.SetRoot(w.getRootView(), true)
	})

	// Center the form
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 10, 1, true).
			AddItem(nil, 0, 1, false), 50, 1, true).
		AddItem(nil, 0, 1, false)

	w.app.SetRoot(modal, true)
	w.app.SetFocus(form)
}

func (w *WifiView) getRootView() tview.Primitive {
	// Navigate up to find the root dashboard
	// This is a simplified approach - in production, store reference to root
	return w.Flex
}

// HandleInput handles keyboard input for the WiFi view.
func (w *WifiView) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 's', 'S':
			// Trigger scan
			if w.device != nil {
				_ = w.adapter.ScanWifi(w.device.Path)
				w.Refresh()
			}
			return nil
		case '/':
			// Focus search
			w.searching = true
			w.app.SetFocus(w.searchInput)
			return nil
		}
	case tcell.KeyEscape:
		if w.searching {
			w.searching = false
			w.app.SetFocus(w.networkList)
			return nil
		}
	}

	return event
}
