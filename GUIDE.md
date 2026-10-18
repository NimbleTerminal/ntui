# `ntui` Developer Guide: Binding & Widget Reference

Because `ntui` is an **Immediate-Mode GUI (IMGUI)**, there are **no widget objects to retain**, **no `OnClick` / `OnChange` callbacks to register**, and **no two-way data synchronization glue**.

Your Go struct fields *are* the source of truth. Every frame inside your `Update` (or UI builder) method, you pass pointers to your state directly into `ntui` widget calls and branch on their `bool` return values using standard Go `if` statements.

---

## 1. The Three Binding Patterns

Every interactive control in `ntui` uses one of three immediate-mode patterns:

### A. Action Trigger (`bool` return)
Push buttons (`Button`, `ButtonSized`) and raw radio buttons (`RadioButton`) return `true` **only on the frame the user activates them** (mouse click-and-release or `Enter`/`Space` when focused):

```go
if ui.Button("Reset Camera") {
    m.cameraYaw = 0
    m.cameraPitch = 0
}
```

### B. Pointer Mutation + Change Notification (`*T` pointer + `bool` return)
Stateful controls (`Checkbox`, `Toggle`, `RadioButtonInt`, `SliderFloat`, `SliderInt`, `TextInput`, `Combo`, `ListBox`, `TabBar`, `ColorEdit`) take a **pointer to your Go variable** (`*bool`, `*int`, `*float64`, `*string`, `*ntui.RGBA`).

1. They read the current value from your pointer to render the control.
2. If the user interacts with the control on this frame, `ntui` mutates your variable **in place** and returns `true` so you can run side effects (like rebuilding a mesh or logging an event):

```go
// 1. Simple binding (mutates m.wireframe automatically when clicked):
ui.Checkbox("Wireframe", &m.wireframe)

// 2. Binding + side effect when the value changes:
if ui.SliderFloat("FOV", &m.fov, 30.0, 120.0, "%.0f°") {
    m.camera.SetFOV(m.fov) // runs only on frames where m.fov changed
}
```

### C. Conditional Scope (`if ui.Begin...` / `End...`)
Containers (`BeginPanel`, `CollapsingHeader`, `TreeNode`) return `true` when open/expanded. Because your UI is just Go code executing top-to-bottom, hiding or collapsing widgets is a normal `if` block—collapsed widgets aren't even evaluated:

```go
if ui.BeginPanel("Inspector", ntui.NewRect(2, 1, 36, 18)) {
    if ui.CollapsingHeader("Transform", true) {
        ui.SliderFloat("Scale", &m.scale, 0.1, 5.0, "%.2fx")
    }
}
ui.EndPanel() // Always call EndPanel() after BeginPanel(), even if collapsed
```

> **Note:** `EndPanel()` is always called unconditionally after `BeginPanel(...)`, whereas `TreePop()` is called **inside** the `if ui.TreeNode(...) { ... ui.TreePop() }` block.

---

## 2. Widget-by-Widget Binding Reference

### Buttons, Checkboxes, Toggles & Radio Buttons

```go
// Push button (3 cells tall by default, or 1 cell tall when Theme.CompactButtons = true)
if ui.Button("Apply Shader") {
    m.compileShader()
}

// Explicit width/height in terminal cells
if ui.ButtonSized("Save", 12, 1) {
    m.save()
}

// Checkbox ([x] / [ ]) bound to *bool
ui.Checkbox("VSync", &m.vsync)

// Pill switch ((─●) / (○─)) bound to *bool
ui.Toggle("Bloom Pass", &m.bloom)

// Integer-bound radio group (mutates m.renderMode to 0, 1, or 2)
ui.BeginHorizontal()
ui.RadioButtonInt("Solid", &m.renderMode, 0)
ui.RadioButtonInt("Wire",  &m.renderMode, 1)
ui.RadioButtonInt("Lit",   &m.renderMode, 2)
ui.EndHorizontal()

// Custom-type radio group using raw RadioButton(label, isActive)
for _, preset := range []string{"Low", "Medium", "Ultra"} {
    if ui.RadioButton(preset, m.qualityPreset == preset) {
        m.qualityPreset = preset
    }
}
```

### Numeric Sliders & Text Input

```go
// Float64 slider with printf format string (drag with mouse or Left/Right arrow keys)
ui.SliderFloat("Exposure", &m.exposure, 0.0, 4.0, "%.2f EV")

// Integer slider
ui.SliderInt("Subdivisions", &m.subdiv, 1, 6)

// Single-line editable text box (width in cells; 0 = fill available width)
if ui.TextInput("Search", &m.filterQuery, 18) {
    m.refreshFilter()
}
```

### Selectors: `TabBar`, `Combo`, and `ListBox`

All three bind to an `*int` index into a `[]string` slice:

```go
// Horizontal TabBar
tabs := []string{"Controls", "Plots", "Trees", "Themes"}
ui.TabBar("MainTabs", &m.activeTab, tabs)

// Inline cycle selector: ◂ Kitty ▸
backends := []string{"Glyph (ntcanvas)", "Raster (Kitty)", "Raster (Sixel)"}
if ui.Combo("Backend", &m.backendIdx, backends) {
    m.switchBackend(m.backendIdx)
}

// Scrollable bordered list box (mouse wheel + Up/Down arrow keys)
meshes := []string{"Torus Knot", "Suzanne", "Utah Teapot", "Stanford Bunny"}
if ui.ListBox("Scene Mesh", &m.selectedMesh, meshes, 6) {
    m.loadMesh(meshes[m.selectedMesh])
}
```

### Color Picker (`ColorEdit`)

Binds directly to an `*ntui.RGBA` struct and renders a live color swatch, hex readout (`#RRGGBB`), and `R`/`G`/`B` sliders:

```go
var lightColor = ntui.RGB(88, 166, 255)

if ui.ColorEdit("Key Light", &m.lightColor) {
    m.scene.SetLightColor(m.lightColor)
}
```

### Hierarchical Trees (`TreeNode` / `TreePop`)

`TreeNode` persists its open/closed state in `ui.Storage` (keyed by widget ID), indents child widgets, and pushes an ID scope automatically when open:

```go
if ui.TreeNode("Scene Root", true) {
    if ui.TreeNode("Rigged Camera", true) {
        ui.SliderFloat("FOV", &m.fov, 30, 120, "%.0f°")
        ui.TreePop()
    }
    if ui.TreeNode("Directional Light", false) {
        ui.Checkbox("Cast Shadows", &m.shadows)
        ui.TreePop()
    }
    ui.TreePop()
}
```

### Read-Only Data, Telemetry & Plots

```go
// Static & formatted text
ui.Label("Renderer Status")
ui.LabelMuted("60.0 Hz target")
ui.Text("Frame: %d (%.2f ms)", m.frameCount, m.frameTimeMs)

// Two-column key/value row (left muted key, right-aligned value)
ui.KeyValue("GPU Memory", "142.4 MB")

// Structured table rows
widths := []int{18, 10, 8}
ui.TableRow([]string{"Pass", "Time", "Draws"}, widths, true)  // header = true
ui.TableRow([]string{"ShadowMap", "0.42 ms", "12"}, widths, false)

// Fractional progress bar in [0.0 .. 1.0]
ui.ProgressBar(m.loadProgress, 0, fmt.Sprintf("Loading %d%%", int(m.loadProgress*100)))

// Inline sparkline & bar histogram (auto-normalized to [min..max] of the slice)
ui.PlotLines("FPS", m.fpsHistory, 0, 2, "60.0 FPS")
ui.PlotHistogram("Frame ms", m.msHistory, 0, 2, "16.6 ms")

// Hover tooltip attached to the immediately preceding widget
ui.Button("Rebuild BVH")
ui.Tooltip("Recomputes bounding volume hierarchy for ray picking")
```

---

## 3. Widget Identity & Loops (`##`, `###`, `PushID`)

`ntui` tracks hover (`Hot`), click/drag (`Active`), and keyboard focus (`Focus`) by hashing each widget's label string into a 64-bit FNV-1a [`ntui.ID`](id.go).

If you create multiple widgets with the **same visible label**, disambiguate their IDs using one of three mechanisms:

### 1. `Display##hidden_suffix`
Text after `##` is hashed into the ID but **hidden from screen**:

```go
ui.Button("Reset##camera") // Displays "Reset", hashes "Reset##camera"
ui.Button("Reset##lights") // Displays "Reset", hashes "Reset##lights"
```

### 2. `Dynamic Text###stable_id`
When a button or panel's visible label changes dynamically (e.g. a live FPS counter in a panel title), `###` tells `ntui` to hash **only** the suffix after `###` so dragging/focus doesn't break when the text updates:

```go
title := fmt.Sprintf("Telemetry (%.1f FPS)###telemetry_panel", m.fps)
if ui.BeginPanel(title, ntui.NewRect(2, 1, 36, 14)) {
    // ...
}
ui.EndPanel()
```

### 3. `PushID` / `PushIDInt` + `PopID` in Loops
When rendering repeated controls in a `for` loop, push the loop index onto the ID stack:

```go
for i := range m.lights {
    ui.PushIDInt(i)
    ui.BeginHorizontal()
    ui.Checkbox("Enabled", &m.lights[i].Enabled)
    ui.SliderFloat("Intensity", &m.lights[i].Intensity, 0, 10, "%.1f")
    ui.EndHorizontal()
    ui.PopID()
}
```

---

## 4. Layout & Positioning Primitives

Inside a panel (or at the root screen cursor), widgets stack **vertically** by default. You can control layout flow inline:

```go
// Place multiple widgets on a single horizontal row
ui.BeginHorizontal()
if ui.Button("Play")  { m.playing = true }
if ui.Button("Pause") { m.playing = false }
ui.EndHorizontal()

// Or append the next widget onto the same line as the previous widget
ui.Checkbox("Grid", &m.showGrid)
ui.SameLine()
ui.Checkbox("Axes", &m.showAxes)

// Indentation and vertical spacing
ui.Indent(2)
ui.LabelMuted("Indented note")
ui.Unindent(2)
ui.Spacing(1)
ui.Separator()
```

---

## 5. Preventing Input Bleed into 3D Viewports (`nt3d`)

When overlaying floating `ntui` panels on top of an interactive 3D viewport (`nt3dwidget`) or custom canvas, check `WantCaptureMouse()` and `WantCaptureKeyboard()` after `EndFrame()` so dragging a slider or typing in a `TextInput` doesn't orbit the 3D camera or trigger global hotkeys:

```go
// 1. Let ntui process the Bubble Tea message and evaluate widgets first:
m.ui.ProcessInput(msg)
m.buildUI() // calls BeginFrame() ... widgets ... EndFrame()

// 2. Only forward mouse/keyboard events to the 3D scene if ntui didn't capture them:
switch msg.(type) {
case tea.MouseMsg:
    if !m.ui.WantCaptureMouse() {
        cmd = m.viewport3D.Update(msg)
    }
case tea.KeyPressMsg:
    if !m.ui.WantCaptureKeyboard() {
        cmd = m.viewport3D.Update(msg)
    }
}
```

---

## 6. Choosing a Render Backend

The same `ui.DrawList` generated between `BeginFrame()` and `EndFrame()` can be rendered by either backend (or switched at runtime):

1. **Glyph Backend (`ntui.GlyphBackend`)**:
   - Renders into an `ntcanvas.GraphemeGrid` using crisp Unicode box drawing (`╭─╮│╰─╯`), eighth-block sliders (`█▉▊▋▌▍▎▏`), and sparklines (`▁▂▃▄▅▆▇█`).
   - Can render standalone (`glyph.RenderToString(ui.DrawList)`) or composite non-empty cells over an existing Bubble Tea view string (`glyph.OverlayString(baseView, ui.DrawList)`).
2. **Raster Backend (`ntui.RasterBackend`)**:
   - Scales cell coordinates by `(CellPixelWidth, CellPixelHeight)` and alpha-blends translucent HUD panels, pixel sliders, and `basicfont` text directly onto an `*image.RGBA` buffer (`raster.RenderToImage(rgbaImg, ui.DrawList)`).
   - Ideal for compositing a HUD directly onto an `nt3d` framebuffer before presenting via `ntpicture` (Kitty / Sixel / iTerm2 graphics protocols).
