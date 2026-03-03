# darknet-tui - Agent Instructions

## Project Overview
A modern, dashboard-style Network Manager TUI for Linux, built with Go and tview. Provides visual WiFi scanning, connection management, ethernet configuration, and hotspot support via NetworkManager D-Bus.

## Tech Stack
- **Language**: Go
- **TUI Framework**: tview
- **Network Integration**: NetworkManager D-Bus (via go-dbus or equivalent)
- **Logging**: zap

## Core Features
1. **Connection Management** - Connect/disconnect/switch between networks
2. **WiFi Scanning** - Live discovery with RSSI visualization
3. **Ethernet Management** - Configure ethernet ports
4. **Hotspot Configuration** - AP/hotspot setup
5. **Command Palette** - Quick command search (Ctrl+P)
6. **Status Bar** - Real-time network status at bottom

## Visual Design
- Dark Blue Theme (matching DarkBlueTUI)
- Signal Gauges - Visual RSSI strength bars (▁▂▃▅▇)
- Connection Indicators - Status dots (●○)
- Progress Bars - For scanning, connecting
- Device Icons - By network type (wifi, ethernet, vpn)
- Color-coded Status - Green (connected), yellow (connecting), gray (disconnected)

## Keybindings
- `Arrow Keys` - Navigate
- `Enter` - Select
- `Tab` - Switch panels
- `q` - Quit
- `/` - Search
- `Ctrl+P` - Command Palette
- `Ctrl+C` - Quit

## Architecture
```
cmd/darknet-tui/
  main.go                 # Entry point, app initialization
internal/
  app/
    app.go                # Main application state & loop
  nmcli/
    adapter.go            # NetworkManager adapter operations
    device.go             # Device operations
    connection.go         # Connection management
  ui/
    dashboard.go          # Main dashboard view
    wifi.go               # WiFi scanning & list
    ethernet.go           # Ethernet management
    hotspot.go            # Hotspot configuration
  widgets/                # Custom tview widgets
pkg/
  theme/
    theme.go              # Color scheme & styles
```

## Error Handling
- Custom error types with error codes
- Wrapped errors with context
- User-friendly error messages in UI

## Testing
- Go testing (stdlib)
- Table-driven tests where appropriate

## Agent Character
**Senior Network Engineer** - Formal, detailed, uses professional network terminology, explains protocols (TCP/IP, DHCP, WPA, etc.), provides comprehensive diagnostics and troubleshooting guidance.

## Commands
- Run: `go run cmd/darknet-tui`
- Build: `go build -o darknet-tui cmd/darknet-tui`
- Test: `go test ./...`
- Lint: `go vet ./...`

## License
GPL-3.0
