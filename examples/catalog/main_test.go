// Copyright (c) 2026 Neomantra Corp.

package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCatalogControlsAndBackendToggle(t *testing.T) {
	m := newCatalogModel()
	defer m.session.Close()
	defer func() { _ = m.pic.Close() }()

	sendKey := func(k string) {
		n, _ := m.Update(tea.KeyPressMsg{Code: rune(k[0]), Text: k})
		m = n.(*catalogModel)
	}

	if !strings.Contains(m.Summary(), "tab=Controls") {
		t.Fatalf("unexpected initial summary: %s", m.Summary())
	}

	sendKey("]")
	if m.activeTab != 1 {
		t.Fatalf("expected activeTab=1 after ']', got %d", m.activeTab)
	}

	sendKey("t")
	if m.themeIdx != 1 {
		t.Fatalf("expected themeIdx=1 after 't', got %d", m.themeIdx)
	}

	sendKey("b")
	if m.borderIdx != 1 {
		t.Fatalf("expected borderIdx=1 after 'b', got %d", m.borderIdx)
	}

	// Toggle Glyph -> Raster -> Glyph and verify keyboard shortcuts remain responsive
	sendKey("g")
	if m.backendMode != 1 {
		t.Fatalf("expected backendMode=1 after 'g', got %d", m.backendMode)
	}
	sendKey("g")
	if m.backendMode != 0 {
		t.Fatalf("expected backendMode=0 after second 'g', got %d", m.backendMode)
	}

	sendKey("]")
	if m.activeTab != 2 {
		t.Fatalf("expected activeTab=2 after ']', got %d", m.activeTab)
	}

	n, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = n.(*catalogModel)
	_ = m.View()
}
