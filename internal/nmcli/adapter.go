// Package nmcli provides NetworkManager D-Bus integration for darknet-tui.
package nmcli

import (
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	// NetworkManager D-Bus constants
	nmBusName           = "org.freedesktop.NetworkManager"
	nmObjectPath        = "/org/freedesktop/NetworkManager"
	nmInterface         = "org.freedesktop.NetworkManager"
	nmSettingsPath      = "/org/freedesktop/NetworkManager/Settings"
	nmSettingsInterface = "org.freedesktop.NetworkManager.Settings"

	// Device types
	DeviceTypeUnknown  DeviceType = 0
	DeviceTypeEthernet DeviceType = 1
	DeviceTypeWifi     DeviceType = 2
	DeviceTypeBT       DeviceType = 5
	DeviceTypeBridge   DeviceType = 13
	DeviceTypeVPN      DeviceType = 14
)

// DeviceType represents NetworkManager device types.
type DeviceType uint32

func (d DeviceType) String() string {
	switch d {
	case DeviceTypeEthernet:
		return "ethernet"
	case DeviceTypeWifi:
		return "wifi"
	case DeviceTypeBT:
		return "bluetooth"
	case DeviceTypeBridge:
		return "bridge"
	case DeviceTypeVPN:
		return "vpn"
	default:
		return "unknown"
	}
}

// NetworkState represents the overall network connectivity state.
type NetworkState uint32

const (
	NetworkStateUnknown      NetworkState = 0
	NetworkStateNone         NetworkState = 10
	NetworkStateDisconnected NetworkState = 20
	NetworkStateConnecting   NetworkState = 40
	NetworkStateConnected    NetworkState = 50
	NetworkStateFull         NetworkState = 70
)

func (n NetworkState) String() string {
	switch n {
	case NetworkStateNone:
		return "none"
	case NetworkStateDisconnected:
		return "disconnected"
	case NetworkStateConnecting:
		return "connecting"
	case NetworkStateConnected:
		return "connected (limited)"
	case NetworkStateFull:
		return "connected"
	default:
		return "unknown"
	}
}

// Adapter represents the NetworkManager D-Bus adapter.
type Adapter struct {
	conn   *dbus.Conn
	obj    dbus.BusObject
	mu     sync.RWMutex
	closed bool
}

// NewAdapter creates a new NetworkManager D-Bus adapter.
func NewAdapter() (*Adapter, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %w", err)
	}

	obj := conn.Object(nmBusName, nmObjectPath)

	return &Adapter{
		conn: conn,
		obj:  obj,
	}, nil
}

// Close closes the D-Bus connection.
func (a *Adapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return nil
	}

	a.closed = true
	return a.conn.Close()
}

// GetState returns the current network connectivity state.
func (a *Adapter) GetState() (NetworkState, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	variant, err := a.obj.GetProperty(nmInterface + ".State")
	if err != nil {
		return NetworkStateUnknown, fmt.Errorf("failed to get network state: %w", err)
	}

	state, ok := variant.Value().(uint32)
	if !ok {
		return NetworkStateUnknown, fmt.Errorf("invalid state type")
	}

	return NetworkState(state), nil
}

// GetVersion returns the NetworkManager version.
func (a *Adapter) GetVersion() (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	variant, err := a.obj.GetProperty(nmInterface + ".Version")
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	version, ok := variant.Value().(string)
	if !ok {
		return "", fmt.Errorf("invalid version type")
	}

	return version, nil
}

// GetDevices returns the list of device object paths.
func (a *Adapter) GetDevices() ([]dbus.ObjectPath, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	variant, err := a.obj.GetProperty(nmInterface + ".Devices")
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	devices, ok := variant.Value().([]dbus.ObjectPath)
	if !ok {
		return nil, fmt.Errorf("invalid devices type")
	}

	return devices, nil
}

// GetActiveConnections returns the list of active connection object paths.
func (a *Adapter) GetActiveConnections() ([]dbus.ObjectPath, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	variant, err := a.obj.GetProperty(nmInterface + ".ActiveConnections")
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}

	connections, ok := variant.Value().([]dbus.ObjectPath)
	if !ok {
		return nil, fmt.Errorf("invalid active connections type")
	}

	return connections, nil
}

// NetworkingEnabled returns whether networking is enabled.
func (a *Adapter) NetworkingEnabled() (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	variant, err := a.obj.GetProperty(nmInterface + ".NetworkingEnabled")
	if err != nil {
		return false, fmt.Errorf("failed to get networking enabled: %w", err)
	}

	enabled, ok := variant.Value().(bool)
	if !ok {
		return false, fmt.Errorf("invalid networking enabled type")
	}

	return enabled, nil
}

// WirelessEnabled returns whether wireless is enabled.
func (a *Adapter) WirelessEnabled() (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	variant, err := a.obj.GetProperty(nmInterface + ".WirelessEnabled")
	if err != nil {
		return false, fmt.Errorf("failed to get wireless enabled: %w", err)
	}

	enabled, ok := variant.Value().(bool)
	if !ok {
		return false, fmt.Errorf("invalid wireless enabled type")
	}

	return enabled, nil
}

// SetWirelessEnabled enables or disables wireless.
func (a *Adapter) SetWirelessEnabled(enabled bool) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	err := a.obj.SetProperty(nmInterface+".WirelessEnabled", dbus.MakeVariant(enabled))
	if err != nil {
		return fmt.Errorf("failed to set wireless enabled: %w", err)
	}

	return nil
}

// GetPrimaryConnection returns the primary connection object path.
func (a *Adapter) GetPrimaryConnection() (dbus.ObjectPath, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	variant, err := a.obj.GetProperty(nmInterface + ".PrimaryConnection")
	if err != nil {
		return "", fmt.Errorf("failed to get primary connection: %w", err)
	}

	conn, ok := variant.Value().(dbus.ObjectPath)
	if !ok {
		return "", fmt.Errorf("invalid primary connection type")
	}

	return conn, nil
}

// Connection returns the D-Bus connection (for use by other packages).
func (a *Adapter) Connection() *dbus.Conn {
	return a.conn
}
