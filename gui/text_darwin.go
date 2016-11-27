package gui

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit
#include "text_darwin.h"
*/
import "C"
import (
	"fmt"
	"image"
	"image/draw"
	"unsafe"
)

func init() {
	typesetter = &darwinTypeSetter{}
}

type darwinTypeSetter struct {
	ctx        C.CGContextRef
	img        *image.RGBA
	glyphCache map[rune]C.CGGlyph
	width      C.CGFloat
	ascent     C.CGFloat
	descent    C.CGFloat
	leading    C.CGFloat
	size       image.Point
}

func newTypeSetter() *darwinTypeSetter {
	return &darwinTypeSetter{}
}

func (t *darwinTypeSetter) SetFont(name string, size float64) error {
	bName := []byte(name)
	var width, ascent, descent, leading C.CGFloat

	if !C.set_font((*C.char)(unsafe.Pointer(&bName[0])), C.CGFloat(size), &width, &ascent, &descent, &leading) {
		return fmt.Errorf("Couldn't set font: %q", name)
	}

	t.width = width
	t.ascent = ascent
	t.descent = descent
	t.leading = leading

	t.size.X = int(t.width + 0.5)
	t.size.Y = int(t.ascent + t.descent + t.leading + 0.5)
	t.img = image.NewRGBA(image.Rect(0, 0, t.size.X, t.size.Y))

	if t.ctx != nil {
		C.CGContextRelease(t.ctx)
	}
	t.ctx = C.create_bitmap_context((*C.uint8_t)(unsafe.Pointer(&t.img.Pix[0])), C.size_t(t.size.X), C.size_t(t.size.Y))

	t.glyphCache = make(map[rune]C.CGGlyph)
	return nil
}

func (t *darwinTypeSetter) CellSize() image.Point {
	return t.size
}

func (t *darwinTypeSetter) getUniCharGlyph(r rune) C.CGGlyph {
	if g, ok := t.glyphCache[r]; ok {
		return g
	}

	var g C.CGGlyph

	if C.get_glyph(C.UniChar(r), &g) {
		t.glyphCache[r] = g
		return g
	}

	if g, ok := t.glyphCache[' ']; ok {
		t.glyphCache[r] = g
		return g
	}

	if C.get_glyph(C.UniChar(' '), &g) {
		t.glyphCache[' '] = g
		t.glyphCache[r] = g
		return g
	}

	return 0
}

func (t *darwinTypeSetter) drawRunes(dst *image.RGBA, x, y int, runes []rune, attr VimTextAttr) {
	rf := float64(attr.Bg.R) / 255
	gf := float64(attr.Bg.G) / 255
	bf := float64(attr.Bg.B) / 255
	af := 1.0

	bg := C.CGColorCreateGenericRGB(C.CGFloat(rf), C.CGFloat(gf), C.CGFloat(bf), C.CGFloat(af))

	rf = float64(attr.Fg.R) / 255
	gf = float64(attr.Fg.G) / 255
	bf = float64(attr.Fg.B) / 255
	af = 1.0

	fg := C.CGColorCreateGenericRGB(C.CGFloat(rf), C.CGFloat(gf), C.CGFloat(bf), C.CGFloat(af))

	for i, r := range runes {
		dx := x + i*t.size.X
		dy := y * t.size.Y
		g := t.getUniCharGlyph(r)
		C.draw_glyph(t.ctx, g, C.int(t.size.X), C.int(t.size.Y), fg, bg)
		draw.Draw(dst, image.Rect(dx, dy, dx+t.size.X, dy+t.size.Y), t.img, image.Point{X: 0, Y: 0}, draw.Over)
	}

	C.CGColorRelease(bg)
	C.CGColorRelease(fg)
}

func (t *darwinTypeSetter) PutChar(dst *image.RGBA, x, y int, c rune, attr VimTextAttr) {
	t.drawRunes(dst, x, y, []rune{c}, attr)
}

func (t *darwinTypeSetter) PutString(dst *image.RGBA, x, y int, s string, attr VimTextAttr) {
	t.drawRunes(dst, x, y, []rune(s), attr)
}
