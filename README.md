# ntui

`ntui` (`nimbleterminal.dev/ntui`) is a pure-Go immediate-mode GUI (IMGUI) library for the NimbleTerminal ecosystem ([Bubble Tea v2](https://github.com/charmbracelet/bubbletea), [`ntcanvas`](https://github.com/NimbleTerminal/ntcanvas), [`ntpicture`](https://github.com/NimbleTerminal/ntpicture), [`ntgpuimage`](https://github.com/NimbleTerminal/ntgpuimage), and [`nt3d`](https://github.com/NimbleTerminal/nt3d)).

Declare controls inline as part of your application's execution flow without managing a retained widget tree. The same UI code evaluates into a backend-agnostic `DrawList` that renders either as terminal cell glyphs (`ntcanvas`) or as pixel-accurate alpha-blended raster graphics (`*image.RGBA` / `ntpicture` / `nt3d` HUD overlays).

## Quick start

```sh
go get nimbleterminal.dev/ntui@latest
```

```go
package main

import (
	tea "charm.land/bubbletea/v2"
	"nimbleterminal.dev/ntui"
)

type model struct {
	ui        *ntui.Context
	glyph     *ntui.GlyphBackend
	wireframe bool
	fov       float64
}

func newModel() *model {
	return &model{
		ui:    ntui.New(ntui.WithScreenSize(80, 24), ntui.WithTheme(ntui.HUDTheme())),
		glyph: ntui.NewGlyphBackend(80, 24),
		fov:   60,
	}
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kp, ok := msg.(tea.KeyPressMsg); ok && kp.String() == "q" && !m.ui.WantCaptureKeyboard() {
		return m, tea.Quit
	}
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.glyph.Resize(ws.Width, ws.Height)
	}

	// 1. Feed terminal events into ntui.
	m.ui.ProcessInput(msg)

	// 2. Evaluate the immediate-mode UI frame and populate the DrawList.
	m.ui.BeginFrame()
	if m.ui.BeginPanel("Camera & Scene", ntui.NewRect(2, 1, 32, 12)) {
		m.ui.Checkbox("Wireframe", &m.wireframe)
		m.ui.SliderFloat("FOV", &m.fov, 30, 120, "%.0f°")
		m.ui.Separator()
		if m.ui.Button("Reset Camera") {
			m.fov = 60
		}
	}
	m.ui.EndPanel()
	m.ui.EndFrame()

	return m, nil
}

func (m *model) View() tea.View {
	v := tea.NewView(m.glyph.RenderToString(m.ui.DrawList))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	return v
}
```

## Architecture

1. **Backend-Agnostic IMGUI Core (`ntui.Context`)**:
   - Tracks `Hot` (hovered), `Active` (held/dragged), and `Focus` (keyboard navigation) widget IDs across frames using 64-bit FNV-1a label hashes (`Display##id` and `Dynamic###stable_id`) and an ID stack (`PushID` / `PopID`).
   - Sequential stack-based layout cursor (`AllocateRect`, `BeginHorizontal` / `EndHorizontal`, `SameLine`, `Indent` / `Unindent`, `BeginPanel` / `EndPanel`).
   - Exposes `WantCaptureMouse()` and `WantCaptureKeyboard()` so 3D viewports (`nt3dwidget`) ignore mouse drags and clicks consumed by floating HUD panels.
2. **Glyph Backend (`ntui.GlyphBackend`)**:
   - Renders `DrawList` commands into an `ntcanvas.GraphemeGrid` using Unicode box-drawing borders (`╭─╮│╰─╯`), fractional eighth-block gauges (`█▉▊▋▌▍▎▏`), and true-color `ntcanvas.CellStyle` attributes, or overlays directly onto an existing terminal view string via `OverlayString`.
3. **Raster Backend (`ntui.RasterBackend`)**:
   - Scales cell coordinates by `(CellPixelWidth, CellPixelHeight)` and alpha-blends translucent panels, crisp borders, continuous pixel sliders, and `golang.org/x/image/font` typography onto `*image.RGBA` or `draw.Image` buffers.

## Widgets & Themes

- **Containers & Navigation**: `BeginPanel` / `EndPanel` (draggable, collapsible, z-ordered floating panels), `CollapsingHeader`, `TabBar`, `TreeNode` / `TreePop`, `Separator`, `BeginHorizontal` / `EndHorizontal`, `SameLine`
- **Controls & Inputs**: `Button`, `Checkbox`, `RadioButton`, `SliderFloat`, `SliderInt`, `InputText`, `Combo`, `ListBox`, `ColorEdit`
- **Data Display & Telemetry**: `Text`, `TextColored`, `KeyValue`, `TableRow`, `ProgressBar`, `PlotLines`, `PlotHistogram`
- **Built-in Themes**: `DefaultTheme()`, `HUDTheme()`, `AmberTheme()`, `CyberpunkTheme()`, `LightTheme()`

## Examples

Run the interactive examples directly with `go run`:

```sh
# UX Element Catalog (4-tab showcase of all widgets, plots, trees, and themes)
go run ./examples/catalog

# Live GPU & Engine Telemetry Dashboard
go run ./examples/dashboard

# Interactive 3D Viewport + Floating HUD Overlay (Glyph & Kitty/Sixel Raster modes)
go run ./examples/hud
```

See [GUIDE.md](GUIDE.md) for the complete widget & state-binding guide, and [DESIGN.md](DESIGN.md) for the technical architecture specification.

## License

Released under the [MIT License](LICENSE), copyright © 2026 Neomantra Corp.
