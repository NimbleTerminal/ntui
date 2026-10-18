// Copyright (c) 2026 Neomantra Corp.

package ntui

import (
	"image/color"

	"nimbleterminal.dev/ntcanvas"
)

// RGBA represents a non-premultiplied 8-bit sRGB color with alpha.
// An alpha of 0 indicates a transparent color (or default terminal color in glyph mode).
type RGBA struct {
	R, G, B, A uint8
}

// RGB constructs an opaque RGBA color.
func RGB(r, g, b uint8) RGBA {
	return RGBA{R: r, G: g, B: b, A: 255}
}

// ColorRGBA constructs an RGBA color with explicit alpha.
func ColorRGBA(r, g, b, a uint8) RGBA {
	return RGBA{R: r, G: g, B: b, A: a}
}

// IsZero reports whether the color is completely transparent / unset.
func (c RGBA) IsZero() bool {
	return c.A == 0
}

// StdRGBA converts c to a standard library color.NRGBA value.
func (c RGBA) StdNRGBA() color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// CanvasColor converts c to an ntcanvas.Color (or zero for default when transparent).
func (c RGBA) CanvasColor() ntcanvas.Color {
	if c.A == 0 {
		return 0
	}
	return ntcanvas.RGBColor(c.R, c.G, c.B)
}

// BorderStyle selects how box outlines are drawn in glyph and raster backends.
type BorderStyle uint8

const (
	BorderNone BorderStyle = iota
	BorderSingle
	BorderRounded
	BorderDouble
	BorderThick
)

// WidgetStyle defines the visual colors and text attributes of a draw primitive.
type WidgetStyle struct {
	FG        RGBA
	BG        RGBA
	Border    RGBA
	Accent    RGBA
	Bold      bool
	Dim       bool
	Underline bool
	Reverse   bool
}

// CellStyle converts WidgetStyle into an ntcanvas.CellStyle.
func (s WidgetStyle) CellStyle() ntcanvas.CellStyle {
	var attr ntcanvas.Attributes
	if s.Bold {
		attr |= ntcanvas.Bold
	}
	if s.Dim {
		attr |= ntcanvas.Dim
	}
	if s.Underline {
		attr |= ntcanvas.Underline
	}
	if s.Reverse {
		attr |= ntcanvas.Reverse
	}
	return ntcanvas.CellStyle{
		Foreground: s.FG.CanvasColor(),
		Background: s.BG.CanvasColor(),
		Attributes: attr,
	}
}

// BorderCellStyle returns the ntcanvas.CellStyle for border runes on this widget.
func (s WidgetStyle) BorderCellStyle() ntcanvas.CellStyle {
	fg := s.Border
	if fg.IsZero() {
		fg = s.FG
	}
	return ntcanvas.CellStyle{
		Foreground: fg.CanvasColor(),
		Background: s.BG.CanvasColor(),
	}
}

// Theme groups semantic widget styles for each interaction state.
type Theme struct {
	Text       WidgetStyle
	Muted      WidgetStyle
	Normal     WidgetStyle
	Hover      WidgetStyle
	Active     WidgetStyle
	Focus      WidgetStyle
	Panel      WidgetStyle
	PanelTitle WidgetStyle
	Input      WidgetStyle
	InputFocus WidgetStyle
	Tooltip    WidgetStyle

	Border      BorderStyle
	PanelBorder BorderStyle

	// CompactButtons controls default Button height: 1 row when true, 3 rows when false.
	CompactButtons bool
}

// DefaultTheme returns a dark Dear ImGui-inspired theme suitable for both
// terminal cell grids and translucent 3D raster HUD overlays.
func DefaultTheme() Theme {
	return Theme{
		Text: WidgetStyle{
			FG: RGB(230, 235, 245),
		},
		Muted: WidgetStyle{
			FG:  RGB(140, 150, 170),
			Dim: true,
		},
		Normal: WidgetStyle{
			FG:     RGB(225, 232, 245),
			BG:     ColorRGBA(42, 54, 76, 230),
			Border: RGB(78, 98, 134),
			Accent: RGB(76, 154, 255),
		},
		Hover: WidgetStyle{
			FG:     RGB(255, 255, 255),
			BG:     ColorRGBA(62, 86, 128, 240),
			Border: RGB(115, 162, 235),
			Accent: RGB(105, 178, 255),
			Bold:   true,
		},
		Active: WidgetStyle{
			FG:     RGB(255, 255, 255),
			BG:     ColorRGBA(38, 108, 196, 250),
			Border: RGB(150, 200, 255),
			Accent: RGB(140, 205, 255),
			Bold:   true,
		},
		Focus: WidgetStyle{
			FG:     RGB(245, 250, 255),
			BG:     ColorRGBA(50, 68, 98, 240),
			Border: RGB(240, 198, 90),
			Accent: RGB(96, 170, 255),
		},
		Panel: WidgetStyle{
			FG:     RGB(220, 226, 238),
			BG:     ColorRGBA(22, 27, 36, 215),
			Border: RGB(64, 78, 104),
			Accent: RGB(76, 154, 255),
		},
		PanelTitle: WidgetStyle{
			FG:     RGB(240, 246, 255),
			BG:     ColorRGBA(36, 50, 76, 235),
			Border: RGB(86, 112, 156),
			Accent: RGB(105, 178, 255),
			Bold:   true,
		},
		Input: WidgetStyle{
			FG:     RGB(230, 236, 248),
			BG:     ColorRGBA(16, 20, 28, 235),
			Border: RGB(70, 86, 116),
			Accent: RGB(76, 154, 255),
		},
		InputFocus: WidgetStyle{
			FG:     RGB(255, 255, 255),
			BG:     ColorRGBA(24, 32, 46, 245),
			Border: RGB(110, 175, 255),
			Accent: RGB(110, 175, 255),
		},
		Tooltip: WidgetStyle{
			FG:     RGB(245, 248, 255),
			BG:     ColorRGBA(14, 18, 24, 245),
			Border: RGB(120, 145, 185),
		},
		Border:         BorderRounded,
		PanelBorder:    BorderRounded,
		CompactButtons: false,
	}
}

// HUDTheme returns a compact translucent theme tailored for floating 3D control panels.
func HUDTheme() Theme {
	t := DefaultTheme()
	t.CompactButtons = true
	t.Panel.BG = ColorRGBA(18, 22, 30, 195)
	return t
}
