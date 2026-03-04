# DarkNetTUI

A modern, dashboard-style Network Manager TUI for Linux.

## Features

- **Dashboard** - Overview of all network connections at a glance
- **WiFi** - Scan, connect, and disconnect from wireless networks
- **Ethernet** - Wired connection status and management
- **Hotspot** - Create and manage a WiFi access point to share your internet connection
- **Command Palette** - Quick access to all actions via Ctrl+P

## Requirements

- Linux with NetworkManager installed and running
- Go 1.25 or later (for building from source)

## Installation

### From Source

```bash
git clone https://github.com/SchumacherFarad/darknet-tui.git
cd darknet-tui
go build -o darknet-tui ./cmd/darknet-tui
sudo cp darknet-tui /usr/local/bin/
```

## Usage

```bash
# Run with default settings
darknet-tui

# Show version
darknet-tui -v

# Custom refresh interval (in seconds)
darknet-tui -refresh 3

# Enable debug logging
darknet-tui -log-level debug -log /tmp/darknet-tui.log
```

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Tab` / `Shift+Tab` | Navigate between panels |
| `Arrow Keys` | Navigate within panels |
| `Enter` | Select / Connect |
| `Space` | Disconnect / Toggle |
| `r` | Refresh current view |
| `Ctrl+p` | Open command palette |
| `Ctrl+c` / `q` | Quit |

### Command Palette Actions

- **Refresh** - Refresh all network data
- **Scan WiFi** - Trigger a WiFi network scan
- **Toggle Wireless** - Enable/disable wireless adapter
- **Quit** - Exit the application

## Hotspot Usage

1. Select **Hotspot** from the device panel (left sidebar)
2. Enter a custom SSID (default: "DarkNet-Hotspot")
3. Enter a password (minimum 8 characters)
4. Click **"Start Hotspot"** to activate
5. Click **"Stop Hotspot"** to deactivate

## Troubleshooting

If you get an error about NetworkManager not running:

```bash
sudo systemctl start NetworkManager
```

For persistent NetworkManager startup:

```bash
sudo systemctl enable NetworkManager
```

## License

GPLv3 - See [LICENSE](LICENSE) for details.
