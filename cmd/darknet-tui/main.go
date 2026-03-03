// darknet-tui - A modern, dashboard-style Network Manager TUI for Linux
//
// Copyright (C) 2024
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/SchumacherFarad/darknet-tui/internal/app"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Parse flags
	var (
		showVersion     bool
		refreshInterval int
		logFile         string
		logLevel        string
	)

	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	flag.IntVar(&refreshInterval, "refresh", 5, "Refresh interval in seconds")
	flag.StringVar(&logFile, "log", "", "Log file path (optional)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	// Show version
	if showVersion {
		fmt.Printf("darknet-tui %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	// Parse log level
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create config
	config := &app.Config{
		RefreshInterval: time.Duration(refreshInterval) * time.Second,
		LogFile:         logFile,
		LogLevel:        level,
	}

	// Create and run application
	application, err := app.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Make sure NetworkManager is running:")
		fmt.Fprintln(os.Stderr, "  sudo systemctl start NetworkManager")
		os.Exit(1)
	}

	// Run the application
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
