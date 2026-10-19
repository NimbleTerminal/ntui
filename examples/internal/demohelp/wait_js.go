// Copyright (c) 2026 Neomantra Corp.
//go:build js && wasm

package demohelp

import (
	"syscall/js"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Tick schedules d through the JavaScript event loop (setTimeout) so continuous
// frame producers under single-threaded WASM always yield to host input and
// render callbacks between frames.
func Tick(d time.Duration, fn func(time.Time) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		setTimeout := js.Global().Get("setTimeout")
		if setTimeout.Type() != js.TypeFunction {
			time.Sleep(max(time.Millisecond, d))
			return fn(time.Now())
		}
		ch := make(chan struct{}, 1)
		cb := js.FuncOf(func(js.Value, []js.Value) any {
			select {
			case ch <- struct{}{}:
			default:
			}
			return nil
		})
		defer cb.Release()
		ms := max(1, int(d/time.Millisecond))
		setTimeout.Invoke(cb, ms)
		<-ch
		return fn(time.Now())
	}
}
