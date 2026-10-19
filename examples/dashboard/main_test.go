// Copyright (c) 2026 Neomantra Corp.

package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestDashboardControls(t *testing.T) {
	m := newDashboardModel()
	defer m.session.Close()
	defer func() { _ = m.pic.Close() }()

	sendKey := func(k string) {
		n, _ := m.Update(tea.KeyPressMsg{Code: rune(k[0]), Text: k})
		m = n.(*dashboardModel)
	}

	sendKey(" ")
	if !m.paused {
		t.Fatalf("expected paused=true after space, got %s", m.Summary())
	}

	sendKey("t")
	if m.themeMode != 1 {
		t.Fatalf("expected themeMode=1 after 't', got %s", m.Summary())
	}

	sendKey("g")
	if m.backendMode != 1 {
		t.Fatalf("expected backendMode=1 after 'g', got %s", m.Summary())
	}
	sendKey("g")
	if m.backendMode != 0 {
		t.Fatalf("expected backendMode=0 after second 'g', got %s", m.Summary())
	}

	sendKey(" ")
	if m.paused {
		t.Fatalf("expected paused=false after second space, got %s", m.Summary())
	}

	n, _ := m.Update(tickMsg(time.Now()))
	m = n.(*dashboardModel)

	n, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = n.(*dashboardModel)
	_ = m.View()
}
