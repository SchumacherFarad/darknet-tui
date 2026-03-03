package nmcli

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const (
	nmDeviceInterface      = "org.freedesktop.NetworkManager.Device"
	nmWirelessInterface    = "org.freedesktop.NetworkManager.Device.Wireless"
	nmWiredInterface       = "org.freedesktop.NetworkManager.Device.Wired"
	nmAccessPointInterface = "org.freedesktop.NetworkManager.AccessPoint"
)

// DeviceState represents NetworkManager device states.
type DeviceState uint32

const (
	DeviceStateUnknown      DeviceState = 0
	DeviceStateUnmanaged    DeviceState = 10
	DeviceStateUnavailable  DeviceState = 20
	DeviceStateDisconnected DeviceState = 30
	DeviceStatePrepare      DeviceState = 40
	DeviceStateConfig       DeviceState = 50
	DeviceStateNeedAuth     DeviceState = 60
	DeviceStateIPConfig     DeviceState = 70
	DeviceStateIPCheck      DeviceState = 80
	DeviceStateSecondaries  DeviceState = 90
	DeviceStateActivated    DeviceState = 100
	DeviceStateDeactivating DeviceState = 110
	DeviceStateFailed       DeviceState = 120
)

func (d DeviceState) String() string {
	switch d {
	case DeviceStateUnmanaged:
		return "unmanaged"
	case DeviceStateUnavailable:
		return "unavailable"
	case DeviceStateDisconnected:
		return "disconnected"
	case DeviceStatePrepare:
		return "preparing"
	case DeviceStateConfig:
		return "configuring"
	case DeviceStateNeedAuth:
		return "need auth"
	case DeviceStateIPConfig:
		return "getting IP"
	case DeviceStateIPCheck:
		return "checking IP"
	case DeviceStateSecondaries:
		return "secondaries"
	case DeviceStateActivated:
		return "connected"
	case DeviceStateDeactivating:
		return "disconnecting"
	case DeviceStateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// IsConnected returns true if the device is connected.
func (d DeviceState) IsConnected() bool {
	return d == DeviceStateActivated
}

// IsConnecting returns true if the device is in the process of connecting.
func (d DeviceState) IsConnecting() bool {
	return d >= DeviceStatePrepare && d < DeviceStateActivated
}

// Device represents a network device.
type Device struct {
	Path      dbus.ObjectPath
	Interface string
	Type      DeviceType
	State     DeviceState
	HWAddress string
	Driver    string
	conn      *dbus.Conn
}

// AccessPoint represents a WiFi access point.
type AccessPoint struct {
	Path     dbus.ObjectPath
	SSID     string
	BSSID    string
	Strength uint8
	Freq     uint32
	Mode     uint32
	Security uint32
}

// GetDeviceInfo retrieves detailed information about a device.
func (a *Adapter) GetDeviceInfo(path dbus.ObjectPath) (*Device, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	obj := a.conn.Object(nmBusName, path)

	device := &Device{
		Path: path,
		conn: a.conn,
	}

	// Get interface name
	if variant, err := obj.GetProperty(nmDeviceInterface + ".Interface"); err == nil {
		if iface, ok := variant.Value().(string); ok {
			device.Interface = iface
		}
	}

	// Get device type
	if variant, err := obj.GetProperty(nmDeviceInterface + ".DeviceType"); err == nil {
		if deviceType, ok := variant.Value().(uint32); ok {
			device.Type = DeviceType(deviceType)
		}
	}

	// Get device state
	if variant, err := obj.GetProperty(nmDeviceInterface + ".State"); err == nil {
		if state, ok := variant.Value().(uint32); ok {
			device.State = DeviceState(state)
		}
	}

	// Get hardware address
	if variant, err := obj.GetProperty(nmDeviceInterface + ".HwAddress"); err == nil {
		if hwAddr, ok := variant.Value().(string); ok {
			device.HWAddress = hwAddr
		}
	}

	// Get driver
	if variant, err := obj.GetProperty(nmDeviceInterface + ".Driver"); err == nil {
		if driver, ok := variant.Value().(string); ok {
			device.Driver = driver
		}
	}

	return device, nil
}

// GetAllDevices returns all network devices with their information.
func (a *Adapter) GetAllDevices() ([]*Device, error) {
	paths, err := a.GetDevices()
	if err != nil {
		return nil, err
	}

	devices := make([]*Device, 0, len(paths))
	for _, path := range paths {
		device, err := a.GetDeviceInfo(path)
		if err != nil {
			continue
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// GetWifiDevices returns only WiFi devices.
func (a *Adapter) GetWifiDevices() ([]*Device, error) {
	devices, err := a.GetAllDevices()
	if err != nil {
		return nil, err
	}

	wifiDevices := make([]*Device, 0)
	for _, d := range devices {
		if d.Type == DeviceTypeWifi {
			wifiDevices = append(wifiDevices, d)
		}
	}

	return wifiDevices, nil
}

// GetEthernetDevices returns only Ethernet devices.
func (a *Adapter) GetEthernetDevices() ([]*Device, error) {
	devices, err := a.GetAllDevices()
	if err != nil {
		return nil, err
	}

	ethDevices := make([]*Device, 0)
	for _, d := range devices {
		if d.Type == DeviceTypeEthernet {
			ethDevices = append(ethDevices, d)
		}
	}

	return ethDevices, nil
}

// ScanWifi triggers a WiFi scan on the specified device.
func (a *Adapter) ScanWifi(devicePath dbus.ObjectPath) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	obj := a.conn.Object(nmBusName, devicePath)
	call := obj.Call(nmWirelessInterface+".RequestScan", 0, map[string]dbus.Variant{})
	if call.Err != nil {
		return fmt.Errorf("failed to request scan: %w", call.Err)
	}

	return nil
}

// GetAccessPoints returns the list of access points for a WiFi device.
func (a *Adapter) GetAccessPoints(devicePath dbus.ObjectPath) ([]*AccessPoint, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	obj := a.conn.Object(nmBusName, devicePath)

	variant, err := obj.GetProperty(nmWirelessInterface + ".AccessPoints")
	if err != nil {
		return nil, fmt.Errorf("failed to get access points: %w", err)
	}

	apPaths, ok := variant.Value().([]dbus.ObjectPath)
	if !ok {
		return nil, fmt.Errorf("invalid access points type")
	}

	accessPoints := make([]*AccessPoint, 0, len(apPaths))
	for _, apPath := range apPaths {
		ap, err := a.getAccessPointInfo(apPath)
		if err != nil {
			continue
		}
		accessPoints = append(accessPoints, ap)
	}

	return accessPoints, nil
}

// getAccessPointInfo retrieves information about an access point.
func (a *Adapter) getAccessPointInfo(path dbus.ObjectPath) (*AccessPoint, error) {
	obj := a.conn.Object(nmBusName, path)

	ap := &AccessPoint{
		Path: path,
	}

	// Get SSID
	if variant, err := obj.GetProperty(nmAccessPointInterface + ".Ssid"); err == nil {
		if ssidBytes, ok := variant.Value().([]byte); ok {
			ap.SSID = string(ssidBytes)
		}
	}

	// Get BSSID (HwAddress)
	if variant, err := obj.GetProperty(nmAccessPointInterface + ".HwAddress"); err == nil {
		if bssid, ok := variant.Value().(string); ok {
			ap.BSSID = bssid
		}
	}

	// Get signal strength
	if variant, err := obj.GetProperty(nmAccessPointInterface + ".Strength"); err == nil {
		if strength, ok := variant.Value().(uint8); ok {
			ap.Strength = strength
		}
	}

	// Get frequency
	if variant, err := obj.GetProperty(nmAccessPointInterface + ".Frequency"); err == nil {
		if freq, ok := variant.Value().(uint32); ok {
			ap.Freq = freq
		}
	}

	// Get mode
	if variant, err := obj.GetProperty(nmAccessPointInterface + ".Mode"); err == nil {
		if mode, ok := variant.Value().(uint32); ok {
			ap.Mode = mode
		}
	}

	// Get security flags
	if variant, err := obj.GetProperty(nmAccessPointInterface + ".WpaFlags"); err == nil {
		if security, ok := variant.Value().(uint32); ok {
			ap.Security = security
		}
	}

	return ap, nil
}

// GetActiveAccessPoint returns the currently connected access point for a WiFi device.
func (a *Adapter) GetActiveAccessPoint(devicePath dbus.ObjectPath) (*AccessPoint, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	obj := a.conn.Object(nmBusName, devicePath)

	variant, err := obj.GetProperty(nmWirelessInterface + ".ActiveAccessPoint")
	if err != nil {
		return nil, fmt.Errorf("failed to get active access point: %w", err)
	}

	apPath, ok := variant.Value().(dbus.ObjectPath)
	if !ok || apPath == "/" {
		return nil, nil // No active access point
	}

	return a.getAccessPointInfo(apPath)
}

// GetCarrier returns whether the ethernet device has carrier (cable connected).
func (a *Adapter) GetCarrier(devicePath dbus.ObjectPath) (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	obj := a.conn.Object(nmBusName, devicePath)

	variant, err := obj.GetProperty(nmWiredInterface + ".Carrier")
	if err != nil {
		return false, fmt.Errorf("failed to get carrier: %w", err)
	}

	carrier, ok := variant.Value().(bool)
	if !ok {
		return false, fmt.Errorf("invalid carrier type")
	}

	return carrier, nil
}

// GetSpeed returns the speed of the ethernet device in Mbps.
func (a *Adapter) GetSpeed(devicePath dbus.ObjectPath) (uint32, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	obj := a.conn.Object(nmBusName, devicePath)

	variant, err := obj.GetProperty(nmWiredInterface + ".Speed")
	if err != nil {
		return 0, fmt.Errorf("failed to get speed: %w", err)
	}

	speed, ok := variant.Value().(uint32)
	if !ok {
		return 0, fmt.Errorf("invalid speed type")
	}

	return speed, nil
}
