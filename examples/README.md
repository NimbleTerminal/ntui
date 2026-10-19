# Examples

Run these from the repository root. Each browser demo publishes its keys through
[go-booba](https://github.com/NimbleMarkets/go-booba) so the same actions appear
as buttons above the terminal.

| Command | What it shows | Controls |
| --- | --- | --- |
| `task run-catalog` | Four-tab UX element catalog (Controls, Plots, Trees & Lists, Themes) + live IMGUI State Inspector | `]`/`[`: switch tab, `t`: cycle theme, `b`: cycle border, `g`: Glyph/Raster, `Tab`: next widget |
| `task run-dashboard` | Live GPU frame-time sparklines, pass breakdown histograms, and simulation controls | `Space`: pause/resume, `t`: cycle theme, `g`: Glyph/Raster, `Tab`: next widget |
| `task run-hud` | Interactive 3D wireframe viewport with draggable floating HUD panels and mouse/keyboard capture | `Space`: pause/orbit, `m`: cycle mesh, `+`/`-` or wheel: zoom, `r`: reset camera, `g`: Glyph/Raster |

Use `q` or `Ctrl+C` to quit. `task serve-wasm-site` builds the WASM demos and
serves all six interactive presets at `http://localhost:8000/web/`.

The `examples` module owns `go-booba` and `ntpicture`. Browser builds select
`wasm.work` and its Bubble Tea replacement so no demo-only dependencies are
added to `ntui`'s root module.
