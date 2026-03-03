package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/SchumacherFarad/darknet-tui/internal/nmcli"
	"github.com/SchumacherFarad/darknet-tui/pkg/theme"
)

// EthernetView displays Ethernet device information and connection management.
type EthernetView struct {
	*tview.Flex
	theme   *theme.Theme
	adapter *nmcli.Adapter
	app     *tview.Application

	// Components
	infoPanel       *tview.TextView
	connectionPanel *tview.List

	// State
	device *nmcli.Device
}

// NewEthernetView creates a new Ethernet view.
func NewEthernetView(app *tview.Application, adapter *nmcli.Adapter, t *theme.Theme) *EthernetView {
	e := &EthernetView{
		Flex:    tview.NewFlex(),
		theme:   t,
		adapter: adapter,
		app:     app,
	}

	e.setupUI()
	return e
}

func (e *EthernetView) setupUI() {
	// Info panel
	e.infoPanel = tview.NewTextView()
	e.infoPanel.SetDynamicColors(true)
	e.infoPanel.SetBackgroundColor(e.theme.Background)
	e.infoPanel.SetTextColor(e.theme.Foreground)
	e.infoPanel.SetBorder(true)
	e.infoPanel.SetTitle(" Device Info ")
	e.infoPanel.SetTitleColor(e.theme.Title)
	e.infoPanel.SetBorderColor(e.theme.Border)

	// Connection panel
	e.connectionPanel = tview.NewList()
	e.connectionPanel.ShowSecondaryText(true)
	e.connectionPanel.SetHighlightFullLine(true)
	e.connectionPanel.SetBackgroundColor(e.theme.Background)
	e.connectionPanel.SetMainTextColor(e.theme.Foreground)
	e.connectionPanel.SetSecondaryTextColor(e.theme.Disconnected)
	e.connectionPanel.SetSelectedBackgroundColor(e.theme.Selection)
	e.connectionPanel.SetSelectedTextColor(e.theme.Foreground)
	e.connectionPanel.SetBorder(true)
	e.connectionPanel.SetTitle(" Saved Connections ")
	e.connectionPanel.SetTitleColor(e.theme.Title)
	e.connectionPanel.SetBorderColor(e.theme.Border)
	e.connectionPanel.SetBorderPadding(0, 0, 1, 1)

	// Layout
	e.Flex.SetDirection(tview.FlexRow)
	e.Flex.AddItem(e.infoPanel, 12, 0, false)
	e.Flex.AddItem(e.connectionPanel, 0, 1, true)

	// Selection handler
	e.connectionPanel.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		e.onConnectionActivated(index)
	})
}

// SetDevice sets the Ethernet device to use.
func (e *EthernetView) SetDevice(device *nmcli.Device) {
	e.device = device
	e.Refresh()
}

// Refresh reloads device information and connections.
func (e *EthernetView) Refresh() {
	if e.device == nil {
		return
	}

	// Get device info
	carrier, _ := e.adapter.GetCarrier(e.device.Path)
	speed, _ := e.adapter.GetSpeed(e.device.Path)

	carrierStatus := "[red]No cable[-]"
	if carrier {
		carrierStatus = "[green]Cable connected[-]"
	}

	speedText := "N/A"
	if speed > 0 {
		speedText = fmt.Sprintf("%d Mbps", speed)
	}

	stateColor := colorToTag(e.theme.Disconnected)
	if e.device.State.IsConnected() {
		stateColor = colorToTag(e.theme.Connected)
	} else if e.device.State.IsConnecting() {
		stateColor = colorToTag(e.theme.Connecting)
	}

	info := fmt.Sprintf(
		"[%s]Interface:[-] %s\n\n"+
			"[%s]MAC Address:[-] %s\n\n"+
			"[%s]Driver:[-] %s\n\n"+
			"[%s]State:[-] [%s]%s[-]\n\n"+
			"[%s]Carrier:[-] %s\n\n"+
			"[%s]Speed:[-] %s",
		colorToTag(e.theme.Primary), e.device.Interface,
		colorToTag(e.theme.Primary), e.device.HWAddress,
		colorToTag(e.theme.Primary), e.device.Driver,
		colorToTag(e.theme.Primary), stateColor, e.device.State.String(),
		colorToTag(e.theme.Primary), carrierStatus,
		colorToTag(e.theme.Primary), speedText,
	)

	e.infoPanel.SetText(info)

	// Get saved ethernet connections
	e.connectionPanel.Clear()

	savedConns, _ := e.adapter.GetSavedConnections()
	activeConns, _ := e.adapter.GetAllActiveConnections()

	// Find active connection for this device
	var activeConnUUID string
	for _, conn := range activeConns {
		for _, devPath := range conn.Devices {
			if devPath == e.device.Path {
				activeConnUUID = conn.UUID
				break
			}
		}
	}

	for _, conn := range savedConns {
		if conn.Type != "802-3-ethernet" {
			continue
		}

		var statusIcon string
		if conn.UUID == activeConnUUID {
			statusIcon = "[green]●[-] "
		} else {
			statusIcon = "[gray]○[-] "
		}

		mainText := fmt.Sprintf("%s%s", statusIcon, conn.ID)
		secondaryText := fmt.Sprintf("  %s", conn.UUID[:8])

		e.connectionPanel.AddItem(mainText, secondaryText, 0, nil)
	}
}

func (e *EthernetView) onConnectionActivated(index int) {
	if e.device == nil {
		return
	}

	// Get saved ethernet connections
	savedConns, _ := e.adapter.GetSavedConnections()
	var ethernetConns []*nmcli.SavedConnection

	for _, conn := range savedConns {
		if conn.Type == "802-3-ethernet" {
			ethernetConns = append(ethernetConns, conn)
		}
	}

	if index < 0 || index >= len(ethernetConns) {
		return
	}

	conn := ethernetConns[index]

	// Check if already connected
	activeConns, _ := e.adapter.GetAllActiveConnections()
	for _, ac := range activeConns {
		if ac.UUID == conn.UUID {
			// Disconnect
			_ = e.adapter.DeactivateConnection(ac.Path)
			e.Refresh()
			return
		}
	}

	// Activate connection
	_, _ = e.adapter.ActivateConnection(conn.Path, e.device.Path)
	e.Refresh()
}

// HandleInput handles keyboard input for the Ethernet view.
func (e *EthernetView) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'r', 'R':
			e.Refresh()
			return nil
		}
	}

	return event
}
