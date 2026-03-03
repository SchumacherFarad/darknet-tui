package nmcli

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	nmActiveConnectionInterface = "org.freedesktop.NetworkManager.Connection.Active"
	nmConnectionSettingsPath    = "/org/freedesktop/NetworkManager/Settings"
	nmConnectionInterface       = "org.freedesktop.NetworkManager.Settings.Connection"
)

// ConnectionState represents the state of an active connection.
type ConnectionState uint32

const (
	ConnectionStateUnknown      ConnectionState = 0
	ConnectionStateActivating   ConnectionState = 1
	ConnectionStateActivated    ConnectionState = 2
	ConnectionStateDeactivating ConnectionState = 3
	ConnectionStateDeactivated  ConnectionState = 4
)

func (c ConnectionState) String() string {
	switch c {
	case ConnectionStateActivating:
		return "activating"
	case ConnectionStateActivated:
		return "activated"
	case ConnectionStateDeactivating:
		return "deactivating"
	case ConnectionStateDeactivated:
		return "deactivated"
	default:
		return "unknown"
	}
}

// ActiveConnection represents an active network connection.
type ActiveConnection struct {
	Path        dbus.ObjectPath
	ID          string
	UUID        string
	Type        string
	State       ConnectionState
	Default     bool
	Default6    bool
	Devices     []dbus.ObjectPath
	Connection  dbus.ObjectPath
	SpecificObj dbus.ObjectPath
}

// SavedConnection represents a saved connection profile.
type SavedConnection struct {
	Path     dbus.ObjectPath
	ID       string
	UUID     string
	Type     string
	Settings map[string]map[string]dbus.Variant
}

// GetActiveConnectionInfo retrieves detailed information about an active connection.
func (a *Adapter) GetActiveConnectionInfo(path dbus.ObjectPath) (*ActiveConnection, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	obj := a.conn.Object(nmBusName, path)

	conn := &ActiveConnection{
		Path: path,
	}

	// Get ID
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".Id"); err == nil {
		if id, ok := variant.Value().(string); ok {
			conn.ID = id
		}
	}

	// Get UUID
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".Uuid"); err == nil {
		if uuid, ok := variant.Value().(string); ok {
			conn.UUID = uuid
		}
	}

	// Get Type
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".Type"); err == nil {
		if connType, ok := variant.Value().(string); ok {
			conn.Type = connType
		}
	}

	// Get State
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".State"); err == nil {
		if state, ok := variant.Value().(uint32); ok {
			conn.State = ConnectionState(state)
		}
	}

	// Get Default (IPv4)
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".Default"); err == nil {
		if def, ok := variant.Value().(bool); ok {
			conn.Default = def
		}
	}

	// Get Default6 (IPv6)
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".Default6"); err == nil {
		if def6, ok := variant.Value().(bool); ok {
			conn.Default6 = def6
		}
	}

	// Get Devices
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".Devices"); err == nil {
		if devices, ok := variant.Value().([]dbus.ObjectPath); ok {
			conn.Devices = devices
		}
	}

	// Get Connection settings path
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".Connection"); err == nil {
		if connPath, ok := variant.Value().(dbus.ObjectPath); ok {
			conn.Connection = connPath
		}
	}

	// Get SpecificObject (e.g., access point for WiFi)
	if variant, err := obj.GetProperty(nmActiveConnectionInterface + ".SpecificObject"); err == nil {
		if specObj, ok := variant.Value().(dbus.ObjectPath); ok {
			conn.SpecificObj = specObj
		}
	}

	return conn, nil
}

// GetAllActiveConnections returns all active connections with their information.
func (a *Adapter) GetAllActiveConnections() ([]*ActiveConnection, error) {
	paths, err := a.GetActiveConnections()
	if err != nil {
		return nil, err
	}

	connections := make([]*ActiveConnection, 0, len(paths))
	for _, path := range paths {
		conn, err := a.GetActiveConnectionInfo(path)
		if err != nil {
			continue
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// ActivateConnection activates a connection on a device.
func (a *Adapter) ActivateConnection(connectionPath, devicePath dbus.ObjectPath) (dbus.ObjectPath, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var activeConnPath dbus.ObjectPath
	specificObject := dbus.ObjectPath("/")

	call := a.obj.Call(nmInterface+".ActivateConnection", 0, connectionPath, devicePath, specificObject)
	if call.Err != nil {
		return "", fmt.Errorf("failed to activate connection: %w", call.Err)
	}

	if err := call.Store(&activeConnPath); err != nil {
		return "", fmt.Errorf("failed to get active connection path: %w", err)
	}

	return activeConnPath, nil
}

// DeactivateConnection deactivates an active connection.
func (a *Adapter) DeactivateConnection(activeConnectionPath dbus.ObjectPath) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	call := a.obj.Call(nmInterface+".DeactivateConnection", 0, activeConnectionPath)
	if call.Err != nil {
		return fmt.Errorf("failed to deactivate connection: %w", call.Err)
	}

	return nil
}

// AddAndActivateConnection creates a new connection and activates it.
// This is useful for connecting to new WiFi networks.
func (a *Adapter) AddAndActivateConnection(
	settings map[string]map[string]dbus.Variant,
	devicePath dbus.ObjectPath,
	specificObject dbus.ObjectPath,
) (dbus.ObjectPath, dbus.ObjectPath, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var connPath, activeConnPath dbus.ObjectPath

	call := a.obj.Call(nmInterface+".AddAndActivateConnection", 0, settings, devicePath, specificObject)
	if call.Err != nil {
		return "", "", fmt.Errorf("failed to add and activate connection: %w", call.Err)
	}

	if err := call.Store(&connPath, &activeConnPath); err != nil {
		return "", "", fmt.Errorf("failed to get connection paths: %w", err)
	}

	return connPath, activeConnPath, nil
}

// ConnectToWifi connects to a WiFi network with the given SSID and password.
func (a *Adapter) ConnectToWifi(devicePath dbus.ObjectPath, apPath dbus.ObjectPath, ssid, password string) error {
	settings := map[string]map[string]dbus.Variant{
		"connection": {
			"type": dbus.MakeVariant("802-11-wireless"),
			"id":   dbus.MakeVariant(ssid),
		},
		"802-11-wireless": {
			"ssid": dbus.MakeVariant([]byte(ssid)),
		},
	}

	if password != "" {
		settings["802-11-wireless-security"] = map[string]dbus.Variant{
			"key-mgmt": dbus.MakeVariant("wpa-psk"),
			"psk":      dbus.MakeVariant(password),
		}
		settings["802-11-wireless"]["security"] = dbus.MakeVariant("802-11-wireless-security")
	}

	_, _, err := a.AddAndActivateConnection(settings, devicePath, apPath)
	return err
}

// GetSavedConnections returns all saved connection profiles.
func (a *Adapter) GetSavedConnections() ([]*SavedConnection, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	settingsObj := a.conn.Object(nmBusName, nmSettingsPath)

	var connPaths []dbus.ObjectPath
	call := settingsObj.Call(nmSettingsInterface+".ListConnections", 0)
	if call.Err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", call.Err)
	}

	if err := call.Store(&connPaths); err != nil {
		return nil, fmt.Errorf("failed to get connection paths: %w", err)
	}

	connections := make([]*SavedConnection, 0, len(connPaths))
	for _, path := range connPaths {
		conn, err := a.getSavedConnectionInfo(path)
		if err != nil {
			continue
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// getSavedConnectionInfo retrieves information about a saved connection.
func (a *Adapter) getSavedConnectionInfo(path dbus.ObjectPath) (*SavedConnection, error) {
	obj := a.conn.Object(nmBusName, path)

	var settings map[string]map[string]dbus.Variant
	call := obj.Call(nmConnectionInterface+".GetSettings", 0)
	if call.Err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", call.Err)
	}

	if err := call.Store(&settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	conn := &SavedConnection{
		Path:     path,
		Settings: settings,
	}

	// Extract common fields
	if connSettings, ok := settings["connection"]; ok {
		if id, ok := connSettings["id"]; ok {
			if idStr, ok := id.Value().(string); ok {
				conn.ID = idStr
			}
		}
		if uuid, ok := connSettings["uuid"]; ok {
			if uuidStr, ok := uuid.Value().(string); ok {
				conn.UUID = uuidStr
			}
		}
		if connType, ok := connSettings["type"]; ok {
			if typeStr, ok := connType.Value().(string); ok {
				conn.Type = typeStr
			}
		}
	}

	return conn, nil
}

// DeleteConnection deletes a saved connection.
func (a *Adapter) DeleteConnection(path dbus.ObjectPath) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	obj := a.conn.Object(nmBusName, path)
	call := obj.Call(nmConnectionInterface+".Delete", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to delete connection: %w", call.Err)
	}

	return nil
}

// CreateHotspot creates a WiFi hotspot.
func (a *Adapter) CreateHotspot(devicePath dbus.ObjectPath, ssid, password string) error {
	settings := map[string]map[string]dbus.Variant{
		"connection": {
			"type":        dbus.MakeVariant("802-11-wireless"),
			"id":          dbus.MakeVariant("Hotspot-" + ssid),
			"autoconnect": dbus.MakeVariant(false),
		},
		"802-11-wireless": {
			"ssid": dbus.MakeVariant([]byte(ssid)),
			"mode": dbus.MakeVariant("ap"),
			"band": dbus.MakeVariant("bg"),
		},
		"802-11-wireless-security": {
			"key-mgmt": dbus.MakeVariant("wpa-psk"),
			"psk":      dbus.MakeVariant(password),
		},
		"ipv4": {
			"method": dbus.MakeVariant("shared"),
		},
		"ipv6": {
			"method": dbus.MakeVariant("ignore"),
		},
	}

	settings["802-11-wireless"]["security"] = dbus.MakeVariant("802-11-wireless-security")

	_, _, err := a.AddAndActivateConnection(settings, devicePath, "/")
	return err
}
