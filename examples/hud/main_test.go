// Copyright (c) 2026 Neomantra Corp.

package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	picture "nimbleterminal.dev/ntpicture"
)

func drain(t *testing.T, m *demoModel, cmd tea.Cmd) []tea.Msg {
	t.Helper()
	var msgs []tea.Msg
	count := 0
	var run func(tea.Cmd)
	run = func(c tea.Cmd) {
		if c == nil {
			return
		}
		count++
		if count > 300 {
			t.Fatal("command loop did not settle")
		}
		msg := c()
		if msg == nil {
			return
		}
		// Do not re-run tickMsg infinitely inside drain
		if _, isTick := msg.(tickMsg); isTick {
			return
		}
		msgs = append(msgs, msg)
		v := reflect.ValueOf(msg)
		if v.Kind() == reflect.Slice && v.Type().Elem() == reflect.TypeOf(tea.Cmd(nil)) {
			for i := 0; i < v.Len(); i++ {
				run(v.Index(i).Interface().(tea.Cmd))
			}
			return
		}
		if _, ok := msg.(tea.RawMsg); ok {
			return
		}
		_, next := m.Update(msg)
		run(next)
	}
	run(cmd)
	return msgs
}

func TestKittyModeSceneUpdatesOnTicks(t *testing.T) {
	m := newDemoModel()
	defer m.session.Close()
	defer func() { _ = m.pic.Close() }()

	drain(t, m, m.session.SetKittyCapability(picture.KittyCapabilitySupported))
	drain(t, m, m.pic.SetMode(picture.PictureKitty))

	// Switch to Raster/Kitty mode by pressing 'g'
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	msgs1 := drain(t, m, cmd)

	st1 := m.pic.Status()
	if st1.Err != nil {
		t.Fatalf("unexpected picture error: %v", st1.Err)
	}
	if st1.Presented == 0 {
		t.Fatalf("expected frame 1 to be presented in Kitty mode, got status=%+v", st1)
	}

	// Simulate 3 animation ticks and verify Presented advances each time
	for i := 0; i < 3; i++ {
		prevPresented := m.pic.Status().Generation
		_, cmd := m.Update(tickMsg(time.Now()))
		msgs := drain(t, m, cmd)
		st := m.pic.Status()
		if st.Presented <= prevPresented {
			t.Fatalf("tick %d did not advance Presented: before=%d after=%+v (msgs=%d)", i, prevPresented, st, len(msgs))
		}
		hasAPC := false
		for _, msg := range msgs {
			if raw, ok := msg.(tea.RawMsg); ok && strings.Contains(fmt.Sprint(raw.Msg), "\x1b_G") {
				hasAPC = true
			}
		}
		if !hasAPC {
			t.Fatalf("tick %d emitted no Kitty APC RawMsg (msgs=%d)", i, len(msgs))
		}
	}
	_ = msgs1
}
