// Copyright (c) 2026 Neomantra Corp.

package ntui

// Layer orders draw commands from back to front regardless of declaration order.
type Layer uint8

const (
	LayerBackground Layer = iota
	LayerWindow
	LayerOverlay
	layerCount
)

// DrawCmdKind identifies the abstract rendering primitive in a DrawCmd.
type DrawCmdKind uint8

const (
	// CmdRect fills an axis-aligned rectangle.
	CmdRect DrawCmdKind = iota
	// CmdBox draws a bordered rectangle with optional interior fill and header title.
	CmdBox
	// CmdText renders single-line text aligned within Bounds.
	CmdText
	// CmdLine draws a horizontal or vertical line between Bounds.Min() and Bounds.Max().
	CmdLine
	// CmdGauge renders a continuous [0..1] progress or slider track with optional thumb and label.
	CmdGauge
)

// DrawCmd is a backend-agnostic rendering instruction emitted during a UI frame.
type DrawCmd struct {
	Kind      DrawCmdKind
	Layer     Layer
	Bounds    Rect
	Clip      Rect
	Text      string
	Align     Align
	Style     WidgetStyle
	Border    BorderStyle
	Fraction  float64
	ShowThumb bool
}

// DrawList collects ordered rendering primitives for a single frame.
type DrawList struct {
	layers    [layerCount][]DrawCmd
	clipStack []Rect
	curLayer  Layer
}

// NewDrawList constructs an empty DrawList with a default unbounded clip rect.
func NewDrawList(screen BoundsProvider) *DrawList {
	dl := &DrawList{}
	dl.Reset(Rect{X: 0, Y: 0, W: 1000, H: 1000})
	if screen != nil {
		dl.Reset(screen.Bounds())
	}
	return dl
}

// BoundsProvider supplies a bounding rectangle.
type BoundsProvider interface {
	Bounds() Rect
}

// Reset clears all recorded commands and resets the clip stack to rootClip.
func (dl *DrawList) Reset(rootClip Rect) {
	for i := range dl.layers {
		dl.layers[i] = dl.layers[i][:0]
	}
	dl.clipStack = append(dl.clipStack[:0], rootClip)
	dl.curLayer = LayerWindow
}

// SetLayer sets the active layer for subsequent Add* calls.
func (dl *DrawList) SetLayer(l Layer) {
	if l < layerCount {
		dl.curLayer = l
	}
}

// Layer returns the currently active layer.
func (dl *DrawList) Layer() Layer {
	return dl.curLayer
}

// PushClipRect intersects r with the current clip rectangle and pushes it onto the stack.
func (dl *DrawList) PushClipRect(r Rect) Rect {
	cur := dl.CurrentClip()
	clipped := cur.Intersect(r)
	dl.clipStack = append(dl.clipStack, clipped)
	return clipped
}

// PushClipRectUnclipped pushes r directly onto the clip stack without intersecting the parent clip.
// This is used for floating panels and tooltips that live in screen space.
func (dl *DrawList) PushClipRectUnclipped(r Rect) Rect {
	dl.clipStack = append(dl.clipStack, r)
	return r
}

// PopClipRect restores the previous clip rectangle on the stack.
func (dl *DrawList) PopClipRect() {
	if len(dl.clipStack) > 1 {
		dl.clipStack = dl.clipStack[:len(dl.clipStack)-1]
	}
}

// CurrentClip returns the active clipping rectangle.
func (dl *DrawList) CurrentClip() Rect {
	if len(dl.clipStack) == 0 {
		return Rect{X: 0, Y: 0, W: 1000, H: 1000}
	}
	return dl.clipStack[len(dl.clipStack)-1]
}

func (dl *DrawList) appendCmd(cmd DrawCmd) {
	cmd.Layer = dl.curLayer
	cmd.Clip = dl.CurrentClip()
	if cmd.Clip.Empty() || cmd.Bounds.Intersect(cmd.Clip).Empty() {
		return
	}
	dl.layers[dl.curLayer] = append(dl.layers[dl.curLayer], cmd)
}

// AddRect emits a filled rectangle primitive.
func (dl *DrawList) AddRect(bounds Rect, style WidgetStyle) {
	dl.appendCmd(DrawCmd{
		Kind:   CmdRect,
		Bounds: bounds,
		Style:  style,
	})
}

// AddBox emits a bordered box primitive with optional title text.
func (dl *DrawList) AddBox(bounds Rect, border BorderStyle, title string, style WidgetStyle) {
	dl.appendCmd(DrawCmd{
		Kind:   CmdBox,
		Bounds: bounds,
		Border: border,
		Text:   title,
		Style:  style,
	})
}

// AddText emits a positioned text primitive aligned within bounds.
func (dl *DrawList) AddText(bounds Rect, text string, align Align, style WidgetStyle) {
	if text == "" || bounds.Empty() {
		return
	}
	dl.appendCmd(DrawCmd{
		Kind:   CmdText,
		Bounds: bounds,
		Text:   text,
		Align:  align,
		Style:  style,
	})
}

// AddTextAt emits left-aligned single-row text starting at p.
func (dl *DrawList) AddTextAt(p Point, width int, text string, style WidgetStyle) {
	if width <= 0 {
		width = len(text)
	}
	dl.AddText(Rect{X: p.X, Y: p.Y, W: width, H: 1}, text, AlignLeft, style)
}

// AddLine emits a horizontal or vertical line primitive occupying bounds.
func (dl *DrawList) AddLine(bounds Rect, style WidgetStyle) {
	dl.appendCmd(DrawCmd{
		Kind:   CmdLine,
		Bounds: bounds,
		Style:  style,
	})
}

// AddGauge emits a fractional [0..1] progress/slider bar primitive.
func (dl *DrawList) AddGauge(bounds Rect, fraction float64, showThumb bool, overlay string, style WidgetStyle) {
	if fraction < 0 {
		fraction = 0
	} else if fraction > 1 {
		fraction = 1
	}
	dl.appendCmd(DrawCmd{
		Kind:      CmdGauge,
		Bounds:    bounds,
		Fraction:  fraction,
		ShowThumb: showThumb,
		Text:      overlay,
		Style:     style,
	})
}

// Commands returns all recorded DrawCmd items ordered from LayerBackground to LayerOverlay.
func (dl *DrawList) Commands() []DrawCmd {
	total := 0
	for i := range dl.layers {
		total += len(dl.layers[i])
	}
	out := make([]DrawCmd, 0, total)
	for i := range dl.layers {
		out = append(out, dl.layers[i]...)
	}
	return out
}

// Len returns the total number of draw commands recorded across all layers.
func (dl *DrawList) Len() int {
	total := 0
	for i := range dl.layers {
		total += len(dl.layers[i])
	}
	return total
}
