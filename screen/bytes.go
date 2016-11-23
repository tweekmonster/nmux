package screen

import "bytes"

// EncodedInt creates 1-5 bytes representing a 32 bit integer depending on the
// value.  This allows for integers to be sent using only the bytes necessary to
// represent their values, instead of using a fixed byte size.
// Each byte contributes 7 bits, with the high bit (0x80) used to indicate that
// the next byte contributes to the value.
func EncodedInt(n int) []byte {
	var b [5]byte
	i := 0
	e := uint32(n)
	if e == 0 {
		return []byte{0}
	}

	for e != 0 {
		b[i] = byte(e)
		e >>= 7
		if e != 0 {
			b[i] |= 0x80
		}
		i++
	}

	return b[:i]
}

// StreamBuffer is the same as bytes.Buffer with a few additions.
type StreamBuffer struct {
	bytes.Buffer
}

func (s *StreamBuffer) WriteOp(op Op) error {
	return s.WriteByte(byte(op))
}

func (s *StreamBuffer) WriteEncodedInt(n int) (int, error) {
	return s.Write(EncodedInt(n))
}

func (s *StreamBuffer) WriteEncodedInts(run ...int) (n int, err error) {
	var in int

	for _, i := range run {
		in, err = s.WriteEncodedInt(i)
		n += in
		if err != nil {
			return
		}
	}

	return
}

func (s *StreamBuffer) WriteColor(c Color) (int, error) {
	return s.Write(EncodedInt(int(c)))
}

// WriteRuneRun writes a run of runes by prefixing the bytes with a length.
func (s *StreamBuffer) WriteRuneRun(str []rune) (n int, err error) {
	n, err = s.WriteEncodedInt(len(str))
	if err != nil {
		return
	}

	var rn int

	for _, r := range str {
		rn, err = s.WriteEncodedInt(int(r))
		n += rn
		if err != nil {
			return
		}
	}

	return
}

// WriteStringRun is a convenience for WriteRuneRun.
func (s *StreamBuffer) WriteStringRun(str string) (n int, err error) {
	return s.WriteRuneRun([]rune(str))
}
