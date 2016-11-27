package gui

import (
	"image"
	"image/draw"
	"io/ioutil"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var typesetter TypeSetter

type TypeSetter interface {
	SetFont(name string, size float64) error
	CellSize() image.Point
	PutChar(dst *image.RGBA, x, y int, c rune, attr VimTextAttr)
	PutString(dst *image.RGBA, x, y int, s string, attr VimTextAttr)
}

// GetTypeSetter returns a typesetter than can render text pixels.  Go's font
// rendering support is not up to snuff.  The default typesetter is good-enough
// as long as TTF fonts are used.
func GetTypeSetter() TypeSetter {
	if typesetter == nil {
		return &defaultTypeSetter{}
	}
	return typesetter
}

type defaultTypeSetter struct {
	f          font.Face
	size       image.Point
	advance    fixed.Int26_6
	ascent     fixed.Int26_6
	lineHeight fixed.Int26_6
}

func (t *defaultTypeSetter) SetFont(name string, size float64) error {
	fontBytes, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	ttfont, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	ttopts := truetype.Options{
		Hinting: font.HintingFull,
		Size:    size,
	}

	t.f = truetype.NewFace(ttfont, &ttopts)
	t.advance, _ = t.f.GlyphAdvance('X')
	m := t.f.Metrics()
	t.lineHeight = m.Descent + m.Ascent
	t.ascent = m.Ascent
	t.size = image.Point{
		X: t.advance.Ceil(),
		Y: t.lineHeight.Ceil(),
	}

	return nil
}

func (t *defaultTypeSetter) CellSize() image.Point {
	return t.size
}

func (t *defaultTypeSetter) PutChar(dst *image.RGBA, x, y int, c rune, attr VimTextAttr) {
	r := image.Rect(0, 0, t.size.X, t.size.Y)
	surface := image.NewRGBA(r)

	for dx := x; dx < x+t.size.X; dx++ {
		for dy := y; dy < y+t.size.Y; dy++ {
			dst.SetRGBA(dx, dy, attr.Bg)
		}
	}

	d := &font.Drawer{
		Dst:  surface,
		Src:  image.NewUniform(attr.Fg),
		Face: t.f,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y*64) + t.ascent},
	}

	d.DrawString(string(c))

	r.Min.X = x * t.size.X
	r.Min.Y = y * t.size.Y
	r.Max.X = r.Min.X + t.size.X
	r.Max.Y = r.Min.Y + t.size.Y
	draw.Draw(dst, r, surface, image.Point{}, draw.Over)
}

func (t *defaultTypeSetter) PutString(dst *image.RGBA, x, y int, s string, attr VimTextAttr) {
	for i, r := range s {
		t.PutChar(dst, x+i, y, r, attr)
	}
}
