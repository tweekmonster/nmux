package gui

import (
	"image/color"

	"github.com/tweekmonster/nmux/screen"
)

type VimTextAttr struct {
	Attrs screen.Attr
	Fg    color.RGBA
	Bg    color.RGBA
	Sp    color.RGBA
}

func CreateVimAttr(a uint8, fg int, bg int, sp int) VimTextAttr {
	return VimTextAttr{
		Attrs: screen.Attr(a),
		Fg:    vimColor(fg),
		Bg:    vimColor(bg),
		Sp:    vimColor(sp),
	}
}

func (v VimTextAttr) Bold() bool {
	return v.Attrs&screen.AttrBold == screen.AttrBold
}

func (v VimTextAttr) Italic() bool {
	return v.Attrs&screen.AttrItalic == screen.AttrItalic
}

func (v VimTextAttr) Underline() bool {
	return v.Attrs&screen.AttrUnderline == screen.AttrUnderline
}

func (v VimTextAttr) Undercurl() bool {
	return v.Attrs&screen.AttrUndercurl == screen.AttrUndercurl
}

func vimColor(rgb int) color.RGBA {
	return color.RGBA{
		R: uint8((rgb >> 16) & 0xff),
		G: uint8((rgb >> 8) & 0xff),
		B: uint8(rgb),
		A: 0xff,
	}
}
