package screen

import "fmt"

const RGBMask = 1 << 31

const (
	ColorForeground uint8 = iota
	ColorBackground
	ColorSpecial
)

var ramp6 = []byte{0x00, 0x5f, 0x87, 0xaf, 0xd7, 0xff}

func cindex(v uint8) uint8 {
	if v < 48 {
		return 0
	}

	if v < 114 {
		return 1
	}

	return (v - 35) / 40
}

func cdist(r1, g1, b1, r2, g2, b2 uint8) int {
	return int((r2-r1)*(r2-r1) + (g2-g1)*(g2-g1) + (b2-b1)*(b2-b1))
}

// Color can represent a 24 bit RGB color or indexed color.
type Color int32

func ColorFromBytes(b []byte) Color {
	_ = b[2]
	return Color((int32(b[0]) << 16) | (int32(b[1]) << 8) | int32(b[2]))
}

func (c Color) Bytes() []byte {
	return []byte{c.R(), c.G(), c.B()}
}

func (c Color) Term() uint8 {
	r, g, b := c.R(), c.G(), c.B()
	qr, qg, qb := cindex(r), cindex(g), cindex(b)
	cr, cg, cb := ramp6[qr], ramp6[qg], ramp6[qb]

	if cr == r && cg == g && cb == b {
		return 16 + 36*qr + 6*qg + qb
	}

	gray := (r + g + b) / 3
	if gray > 238 {
		gray = 23
	} else {
		gray = (gray - 3) / 10
	}

	d := cdist(cr, cg, cb, r, g, b)
	if cdist(gray, gray, gray, r, g, b) < d {
		return 232 + gray
	}
	return 16 + 36*qr + 6*qg + qb
}

func (c Color) R() uint8 {
	return uint8((c >> 16) & 0xffff)
}

func (c Color) G() uint8 {
	return uint8((c >> 8) & 0xff)
}

func (c Color) B() uint8 {
	return uint8(c & 0xff)
}

func (c Color) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R(), c.G(), c.B())
}
