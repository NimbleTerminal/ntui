//go:build integration

package ntuiebiten

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync/atomic"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
	"nimbleterminal.dev/ntui"
)

type gpuCall struct {
	fn   func()
	done chan struct{}
}

var calls = make(chan gpuCall)
var stopped atomic.Bool

type testGame struct{}

func (*testGame) Update() error {
	if stopped.Load() {
		return ebiten.Termination
	}
	select {
	case call := <-calls:
		call.fn()
		close(call.done)
	default:
	}
	return nil
}
func (*testGame) Draw(*ebiten.Image)         {}
func (*testGame) Layout(int, int) (int, int) { return 1, 1 }
func TestMain(m *testing.M) {
	result := make(chan int, 1)
	go func() { code := m.Run(); stopped.Store(true); result <- code }()
	ebiten.SetWindowVisible(false)
	ebiten.SetRunnableOnUnfocused(true)
	err := ebiten.RunGame(&testGame{})
	code := <-result
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		code = 1
	}
	os.Exit(code)
}
func onGPU(fn func()) { done := make(chan struct{}); calls <- gpuCall{fn, done}; <-done }
func pixels(img *ebiten.Image) *image.RGBA {
	out := image.NewRGBA(img.Bounds())
	img.ReadPixels(out.Pix)
	return out
}

var testStyle = ntui.WidgetStyle{FG: ntui.ColorRGBA(230, 240, 250, 190), BG: ntui.ColorRGBA(80, 120, 180, 110), Border: ntui.ColorRGBA(180, 220, 240, 200), Accent: ntui.ColorRGBA(240, 170, 50, 180)}

func command(kind ntui.DrawCmdKind) ntui.DrawCmd {
	return ntui.DrawCmd{Kind: kind, Bounds: ntui.Rect{X: 1, Y: 1, W: 15, H: 3}, Clip: ntui.Rect{W: 30, H: 10}, Style: testStyle, Border: ntui.BorderRounded, Text: "Hi ✓ ▾", Align: ntui.AlignRight, Fraction: .6, ShowThumb: true, Values: []float64{.1, .9, .4, .7, .2}}
}
func TestAllCommandsClippingPreservesLayout(t *testing.T) {
	for _, kind := range []ntui.DrawCmdKind{ntui.CmdRect, ntui.CmdBox, ntui.CmdText, ntui.CmdLine, ntui.CmdGauge, ntui.CmdSparkline} {
		for _, hist := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/hist=%v", kind, hist), func(t *testing.T) {
				var full, clipped *image.RGBA
				onGPU(func() {
					b := New()
					defer b.Dispose()
					a := ebiten.NewImage(240, 100)
					defer a.Dispose()
					d := ebiten.NewImage(240, 100)
					defer d.Dispose()
					c := command(kind)
					c.Histogram = hist
					options := RenderOptions{ScaleX: 9.5, ScaleY: 17.25, OffsetX: 3.25, OffsetY: 2.5}
					b.Render(a, []ntui.DrawCmd{c}, options)
					c.Clip = ntui.Rect{X: 6, Y: 2, W: 6, H: 2}
					b.Render(d, []ntui.DrawCmd{c}, options)
					full, clipped = pixels(a), pixels(d)
				})
				// Absolute rounded cell edges for the clip, independently specified.
				crop := image.Rect(60, 37, 117, 72)
				ink := 0
				for y := 0; y < 100; y++ {
					for x := 0; x < 240; x++ {
						got := clipped.RGBAAt(x, y)
						want := color.RGBA{}
						if image.Pt(x, y).In(crop) {
							want = full.RGBAAt(x, y)
						}
						if got != want {
							t.Fatalf("clipping changed layout at %d,%d: %v != %v", x, y, got, want)
						}
						if full.RGBAAt(x, y).A > 0 {
							ink++
						}
					}
				}
				if ink == 0 {
					t.Fatal("command rendered no pixels")
				}
			})
		}
	}
}
func TestFractionalTilingSubimageAndOrder(t *testing.T) {
	var got *image.RGBA
	onGPU(func() {
		b := New()
		defer b.Dispose()
		dst := ebiten.NewImage(60, 60)
		defer dst.Dispose()
		sub := dst.SubImage(image.Rect(11, 13, 49, 48)).(*ebiten.Image)
		var cmds []ntui.DrawCmd
		for x := 0; x < 10; x++ {
			cmds = append(cmds, ntui.DrawCmd{Kind: ntui.CmdRect, Bounds: ntui.Rect{X: x, W: 1, H: 4}, Clip: ntui.Rect{W: 10, H: 4}, Style: ntui.WidgetStyle{BG: ntui.RGB(255, 0, 0)}})
		}
		cmds = append(cmds, ntui.DrawCmd{Kind: ntui.CmdRect, Layer: ntui.LayerBackground, Bounds: ntui.Rect{X: 4, W: 2, H: 4}, Clip: ntui.Rect{W: 10, H: 4}, Style: ntui.WidgetStyle{BG: ntui.RGB(0, 255, 0)}})
		b.Render(sub, cmds, RenderOptions{ScaleX: 2.5, ScaleY: 3.25, OffsetX: .5, OffsetY: 1.25})
		got = pixels(dst)
	})
	for y := 0; y < 60; y++ {
		for x := 0; x < 60; x++ {
			want := color.RGBA{}
			if image.Pt(x, y).In(image.Rect(12, 14, 37, 27)) {
				want = color.RGBA{255, 0, 0, 255}
				if x >= 22 && x < 27 {
					want = color.RGBA{0, 255, 0, 255}
				}
			}
			if got.RGBAAt(x, y) != want {
				t.Fatalf("pixel %d,%d: %v != %v", x, y, got.RGBAAt(x, y), want)
			}
		}
	}
}
func TestStraightAlphaPanelsAndGlyphEdges(t *testing.T) {
	for _, background := range []uint8{0, 245} {
		var got, mask *image.RGBA
		onGPU(func() {
			b := New()
			defer b.Dispose()
			a := ebiten.NewImage(160, 60)
			defer a.Dispose()
			m := ebiten.NewImage(160, 60)
			defer m.Dispose()
			a.Fill(color.RGBA{background, background, background, 255})
			c := ntui.DrawCmd{Kind: ntui.CmdRect, Bounds: ntui.Rect{W: 16, H: 1}, Clip: ntui.Rect{W: 16, H: 4}, Style: ntui.WidgetStyle{BG: ntui.ColorRGBA(200, 80, 30, 128)}}
			b.Render(a, []ntui.DrawCmd{c}, RenderOptions{})
			c.Kind = ntui.CmdText
			c.Bounds.Y = 1
			c.Text = "Aa ✓ ▾"
			c.Style.FG = ntui.ColorRGBA(200, 80, 30, 128)
			b.Render(a, []ntui.DrawCmd{c}, RenderOptions{})
			c.Style.FG = ntui.RGB(255, 255, 255)
			b.Render(m, []ntui.DrawCmd{c}, RenderOptions{})
			got, mask = pixels(a), pixels(m)
		})
		check := func(p color.RGBA, alpha float64) {
			for i, v := range []uint8{p.R, p.G, p.B} {
				src := []float64{200, 80, 30}[i]
				expected := src*alpha + float64(background)*(1-alpha)
				if math.Abs(float64(v)-expected) > 3 {
					t.Errorf("background %d got %v expected channel %d ~ %.1f", background, p, i, expected)
					return
				}
			}
		}
		check(got.RGBAAt(4, 4), 128.0/255)
		edges := 0
		for y := 18; y < 36; y++ {
			for x := 0; x < 144; x++ {
				a := mask.RGBAAt(x, y).A
				if a > 5 && a < 250 {
					check(got.RGBAAt(x, y), float64(a)/255*128/255)
					edges++
				}
			}
		}
		if edges == 0 {
			t.Fatal("no antialiased glyph edges tested")
		}
	}
}
func TestMeasuredTextSymbolsAndCleanup(t *testing.T) {
	var imgs [3]*image.RGBA
	var width float64
	var failure string
	onGPU(func() {
		source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
		if err != nil {
			failure = err.Error()
			return
		}
		face := &text.GoTextFace{Source: source, Size: 17}
		b := New(WithFontFaces(face, nil))
		_, width = textRuns("Wi ✓ ▾", face)
		for i, align := range []ntui.Align{ntui.AlignLeft, ntui.AlignCenter, ntui.AlignRight} {
			dst := ebiten.NewImage(200, 40)
			c := ntui.DrawCmd{Kind: ntui.CmdText, Bounds: ntui.Rect{W: 20, H: 2}, Clip: ntui.Rect{W: 20, H: 2}, Text: "Wi ✓ ▾", Align: align, Style: ntui.WidgetStyle{FG: ntui.RGB(255, 255, 255)}}
			b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{ScaleX: 10, ScaleY: 20})
			imgs[i] = pixels(dst)
			if b.regular != face {
				failure = "font was recreated"
			}
			if i == 2 {
				b.Dispose()
				b.Dispose()
				b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{})
				if !bytes.Equal(imgs[i].Pix, pixels(dst).Pix) {
					failure = "disposed backend drew"
				}
				dst.Fill(color.White)
				if pixels(dst).RGBAAt(0, 0).A != 255 {
					failure = "caller destination disposed"
				}
			}
			dst.Dispose()
		}
	})
	if failure != "" {
		t.Fatal(failure)
	}
	minX := func(img *image.RGBA) int {
		x0 := 200
		for y := 0; y < 40; y++ {
			for x := 0; x < 200; x++ {
				if img.RGBAAt(x, y).A > 32 {
					x0 = min(x0, x)
				}
			}
		}
		return x0
	}
	left := minX(imgs[0])
	for i := 1; i < 3; i++ {
		expected := (200 - width) * float64(i) / 2
		shift := float64(minX(imgs[i]) - left)
		if math.Abs(shift-expected) > 2 {
			t.Fatalf("alignment %d shift %.1f expected %.1f", i, shift, expected)
		}
	}
	// Every fallback symbol must draw visible ink even in a font lacking it.
	onGPU(func() {
		b := New()
		defer b.Dispose()
		dst := ebiten.NewImage(32, 32)
		defer dst.Dispose()
		for _, r := range symbols {
			dst.Clear()
			b.Render(dst, []ntui.DrawCmd{{Kind: ntui.CmdText, Bounds: ntui.Rect{W: 3, H: 1}, Clip: ntui.Rect{W: 3, H: 1}, Text: string(r), Style: ntui.WidgetStyle{FG: ntui.RGB(255, 255, 255)}}}, RenderOptions{})
			p := pixels(dst)
			ink := false
			for i := 3; i < len(p.Pix); i += 4 {
				ink = ink || p.Pix[i] > 0
			}
			if !ink {
				failure = fmt.Sprintf("missing symbol %c", r)
			}
		}
	})
	if failure != "" {
		t.Fatal(failure)
	}
}
func TestVisualContactSheet(t *testing.T) {
	path := os.Getenv("NTUI_TEST_IMAGE")
	if path == "" {
		t.Skip("set NTUI_TEST_IMAGE to export a visual review PNG")
	}
	var img *image.RGBA
	onGPU(func() {
		b := New()
		defer b.Dispose()
		dst := ebiten.NewImage(700, 430)
		defer dst.Dispose()
		dst.Fill(color.RGBA{20, 24, 32, 255})
		vectorBG := ntui.DrawCmd{Kind: ntui.CmdRect, Bounds: ntui.Rect{X: 38, W: 40, H: 24}, Clip: ntui.Rect{W: 80, H: 24}, Style: ntui.WidgetStyle{BG: ntui.RGB(240, 242, 246)}}
		b.Render(dst, []ntui.DrawCmd{vectorBG}, RenderOptions{})
		for col := 0; col < 2; col++ {
			for i := 0; i < 7; i++ {
				kind := ntui.DrawCmdKind(min(i, 5))
				c := command(kind)
				c.Bounds = ntui.Rect{X: 2 + col*38, Y: 1 + i*3, W: 32, H: 3}
				c.Clip = ntui.Rect{W: 80, H: 24}
				c.Histogram = i == 6
				c.Text = "ntui  ✓  ▾  ◂  ●  ○"
				b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{})
			}
		}
		img = pixels(dst)
	})
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestEmptyClippedAndInvalidCommands(t *testing.T) {
	var got *image.RGBA
	onGPU(func() {
		b := New()
		defer b.Dispose()
		dst := ebiten.NewImage(30, 30)
		defer dst.Dispose()
		c := command(ntui.CmdRect)
		b.Render(nil, []ntui.DrawCmd{c}, RenderOptions{})
		b.Render(dst, nil, RenderOptions{})
		b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{ScaleX: math.NaN()})
		b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{ScaleY: -1})
		b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{OffsetX: math.Inf(1)})
		c.Clip = ntui.Rect{}
		b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{})
		c.Clip = ntui.Rect{X: 50, Y: 50, W: 10, H: 10}
		b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{})
		c.Clip = ntui.Rect{W: 10, H: 10}
		c.Bounds = ntui.Rect{}
		b.Render(dst, []ntui.DrawCmd{c}, RenderOptions{})
		got = pixels(dst)
	})
	for _, v := range got.Pix {
		if v != 0 {
			t.Fatal("empty or invalid command drew pixels")
		}
	}
}
