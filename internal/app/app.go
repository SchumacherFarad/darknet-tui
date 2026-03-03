// Package app provides the main application logic for darknet-tui.
package app

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/SchumacherFarad/darknet-tui/internal/nmcli"
	"github.com/SchumacherFarad/darknet-tui/internal/ui"
	"github.com/SchumacherFarad/darknet-tui/internal/widgets"
	"github.com/SchumacherFarad/darknet-tui/pkg/theme"
)

// App is the main application struct.
type App struct {
	tviewApp  *tview.Application
	adapter   *nmcli.Adapter
	theme     *theme.Theme
	logger    *zap.Logger
	dashboard *ui.Dashboard

	// Command palette
	commandPalette *widgets.CommandPalette
	commands       []widgets.Command
	paletteVisible bool

	// State
	refreshTicker *time.Ticker
	done          chan struct{}
}

// Config holds application configuration.
type Config struct {
	RefreshInterval time.Duration
	LogFile         string
	LogLevel        zapcore.Level
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		RefreshInterval: 5 * time.Second,
		LogFile:         "",
		LogLevel:        zapcore.InfoLevel,
	}
}

// New creates a new application instance.
func New(config *Config) (*App, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Setup logger
	logger, err := setupLogger(config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logger: %w", err)
	}

	// Create NetworkManager adapter
	adapter, err := nmcli.NewAdapter()
	if err != nil {
		logger.Error("Failed to connect to NetworkManager", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to NetworkManager: %w", err)
	}

	// Check NetworkManager version
	version, err := adapter.GetVersion()
	if err != nil {
		logger.Warn("Could not get NetworkManager version", zap.Error(err))
	} else {
		logger.Info("Connected to NetworkManager", zap.String("version", version))
	}

	// Create tview application
	tviewApp := tview.NewApplication()

	// Create theme
	t := theme.DarkBlue()

	app := &App{
		tviewApp: tviewApp,
		adapter:  adapter,
		theme:    t,
		logger:   logger,
		done:     make(chan struct{}),
	}

	// Setup UI
	app.setupUI()
	app.setupCommands()

	// Setup refresh ticker
	if config.RefreshInterval > 0 {
		app.refreshTicker = time.NewTicker(config.RefreshInterval)
		go app.refreshLoop()
	}

	return app, nil
}

func setupLogger(config *Config) (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var core zapcore.Core

	if config.LogFile != "" {
		// Log to file
		file, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(file),
			config.LogLevel,
		)
	} else {
		// Discard logs in TUI mode (or log to a temp file)
		core = zapcore.NewNopCore()
	}

	return zap.New(core), nil
}

func (a *App) setupUI() {
	// Create dashboard
	a.dashboard = ui.NewDashboard(a.tviewApp, a.adapter, a.theme)

	// Create command palette
	a.commandPalette = widgets.NewCommandPalette(a.theme)

	// Set root
	a.tviewApp.SetRoot(a.dashboard, true)

	// Setup input handler
	a.tviewApp.SetInputCapture(a.handleInput)
}

func (a *App) setupCommands() {
	a.commands = []widgets.Command{
		{
			Name:        "Refresh",
			Description: "Refresh all network data",
			Action:      func() { a.dashboard.Refresh() },
		},
		{
			Name:        "Scan WiFi",
			Description: "Trigger WiFi network scan",
			Action:      func() { a.dashboard.Refresh() },
		},
		{
			Name:        "Toggle Wireless",
			Description: "Enable/Disable wireless",
			Action:      a.toggleWireless,
		},
		{
			Name:        "Quit",
			Description: "Exit the application",
			Action:      func() { a.tviewApp.Stop() },
		},
	}

	a.commandPalette.SetCommands(a.commands)
}

func (a *App) handleInput(event *tcell.EventKey) *tcell.EventKey {
	// Handle command palette toggle
	if event.Key() == tcell.KeyCtrlP {
		a.toggleCommandPalette()
		return nil
	}

	// If command palette is visible, handle its input
	if a.paletteVisible {
		switch event.Key() {
		case tcell.KeyEscape:
			a.hideCommandPalette()
			return nil
		case tcell.KeyEnter:
			// Execute selected command
			a.executeSelectedCommand()
			return nil
		}
		return event
	}

	// Pass to dashboard
	return a.dashboard.HandleInput(event)
}

func (a *App) toggleCommandPalette() {
	if a.paletteVisible {
		a.hideCommandPalette()
	} else {
		a.showCommandPalette()
	}
}

func (a *App) showCommandPalette() {
	a.paletteVisible = true

	// Create command list
	commandList := tview.NewList()
	commandList.ShowSecondaryText(true)
	commandList.SetHighlightFullLine(true)
	commandList.SetBackgroundColor(a.theme.StatusBar)
	commandList.SetMainTextColor(a.theme.Foreground)
	commandList.SetSecondaryTextColor(a.theme.Disconnected)
	commandList.SetSelectedBackgroundColor(a.theme.Selection)
	commandList.SetSelectedTextColor(a.theme.Foreground)

	for _, cmd := range a.commands {
		commandList.AddItem(cmd.Name, cmd.Description, 0, nil)
	}

	commandList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(a.commands) {
			a.hideCommandPalette()
			a.commands[index].Action()
		}
	})

	// Create modal container
	modal := tview.NewFlex().SetDirection(tview.FlexRow)
	modal.SetBackgroundColor(a.theme.StatusBar)
	modal.SetBorder(true)
	modal.SetTitle(" Commands (Ctrl+P) ")
	modal.SetTitleColor(a.theme.Title)
	modal.SetBorderColor(a.theme.BorderFocus)

	modal.AddItem(a.commandPalette, 1, 0, true)
	modal.AddItem(commandList, 0, 1, true)

	// Center the modal
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(modal, 15, 1, true).
			AddItem(nil, 0, 2, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	// Create pages to overlay
	pages := tview.NewPages()
	pages.AddPage("dashboard", a.dashboard, true, true)
	pages.AddPage("palette", centered, true, true)

	a.tviewApp.SetRoot(pages, true)
	a.tviewApp.SetFocus(commandList)
}

func (a *App) hideCommandPalette() {
	a.paletteVisible = false
	a.tviewApp.SetRoot(a.dashboard, true)
	a.tviewApp.SetFocus(a.dashboard.GetFocusable())
}

func (a *App) executeSelectedCommand() {
	// This would be called when Enter is pressed on a command
	a.hideCommandPalette()
}

func (a *App) toggleWireless() {
	enabled, err := a.adapter.WirelessEnabled()
	if err != nil {
		a.logger.Error("Failed to get wireless state", zap.Error(err))
		return
	}

	if err := a.adapter.SetWirelessEnabled(!enabled); err != nil {
		a.logger.Error("Failed to toggle wireless", zap.Error(err))
		return
	}

	a.dashboard.Refresh()
}

func (a *App) refreshLoop() {
	for {
		select {
		case <-a.refreshTicker.C:
			a.tviewApp.QueueUpdateDraw(func() {
				a.dashboard.Refresh()
			})
		case <-a.done:
			return
		}
	}
}

// Run starts the application.
func (a *App) Run() error {
	a.logger.Info("Starting darknet-tui")
	return a.tviewApp.Run()
}

// Stop stops the application.
func (a *App) Stop() {
	a.logger.Info("Stopping darknet-tui")

	// Stop refresh ticker
	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
	}

	// Signal done
	close(a.done)

	// Close adapter
	if a.adapter != nil {
		_ = a.adapter.Close()
	}

	// Stop tview
	a.tviewApp.Stop()
}
