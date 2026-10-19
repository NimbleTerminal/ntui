// Copyright (c) 2026 Neomantra Corp.
//go:build !js || !wasm

package demohelp

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// Tick schedules a periodic Bubble Tea tick command.
func Tick(d time.Duration, fn func(time.Time) tea.Msg) tea.Cmd {
	return tea.Tick(d, fn)
}
