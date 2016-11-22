package screen

// TODO: There should be a buffer that's managed elsewhere that this can
// unconditionally write to.

func (s *Screen) writeByte(c byte) error {
	return s.payload.WriteByte(c)
}

func (s *Screen) writeRune(r rune) (int, error) {
	return s.payload.WriteRune(r)
}

// encodedInt creates 1-5 bytes representing a 32 bit integer depending on the
// value.  This allows for integers to be sent using only the bytes necessary to
// represent their values, instead of using a fixed byte size.
// Each byte contributes 7 bits, with the high bit (0x80) used to indicate that
// the next byte contributes to the value.
func (s *Screen) encodedInt(n int) []byte {
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

func (s *Screen) writeEncodedInt(n int) (int, error) {
	return s.write(s.encodedInt(n))
}

func (s *Screen) write(p []byte) (int, error) {
	return s.payload.Write(p)
}

// writeRange sends a range of characters to render.  The ranges *must* have the
// same display attributes.  A run of the same character will be sent as a
// "condensed" operation.
func (s *Screen) writeRange(i1, i2 int) {
	run := make([]rune, 0, i2-i1)
	runs := make([][2]int, 0, i2-i1)

	firstC := s.Buffer[i1].Char
	runStart := 0
	runEnd := 0

	for i, c := range s.Buffer[i1:i2] {
		if c.Char == firstC {
			runEnd++
		} else {
			runs = append(runs, [2]int{runStart, runEnd})
			firstC = c.Char
			runStart = i
			runEnd = i + 1
		}
		run = append(run, c.Char)
	}

	if runEnd-runStart > 0 {
		runs = append(runs, [2]int{runStart, runEnd})
	}

	send := make([]rune, i2-i1)
	i := i1
	j := 0

	for _, r := range runs {
		if r[1]-r[0] < 2 {
			// Run is too short.
			copy(send[j:], run[r[0]:r[1]])
			j += r[1] - r[0]
		} else {
			if j > 0 {
				// Flush pending characters.
				s.writeByte(byte(OpPut))
				s.writeEncodedInt(i)
				s.writeEncodedInt(j)
				i += j

				for _, c := range send[:j] {
					s.writeEncodedInt(int(c))
				}
			}

			// Getting here means that the run is long enough to send as a condensed
			// render operation.
			s.writeByte(byte(OpPutRep))
			s.writeEncodedInt(i)
			s.writeEncodedInt(r[1] - r[0])
			s.writeEncodedInt(int(run[r[0]]))

			i += r[1] - r[0]
			j = 0
		}
	}

	if j > 0 {
		s.writeByte(byte(OpPut))
		s.writeEncodedInt(i)
		s.writeEncodedInt(j)
		for _, c := range send[:j] {
			s.writeEncodedInt(int(c))
		}
	}
}

func (s *Screen) writeScroll(delta int) {
	s.writeByte(byte(OpScroll))
	s.write(s.DefaultAttrs.Bg.Bytes())
	s.write([]byte{byte(delta >> 8), byte(delta)})
	s.writeEncodedInt(s.scroll.tl.Y)
	s.writeEncodedInt(s.scroll.br.Y)
	s.writeEncodedInt(s.scroll.tl.X)
	s.writeEncodedInt(s.scroll.br.X)
}

func (s *Screen) writeClear() {
	// Clear the attribute counter of colors that are no longer used in the cells.
	for attr := range s.attrCounter {
		if attr.id != 0 {
			if attr != s.CurAttrs {
				delete(s.attrCounter, attr)
			}

			if _, ok := s.sentAttrs[attr]; ok {
				delete(s.sentAttrs, attr)
			}
		}
	}

	s.lastSent = nil

	s.writeByte(byte(OpClear))

	attr := s.DefaultAttrs
	// Clear is a special case that receives the default colors before resetting
	// the palette on the client side.
	s.writeEncodedInt(int(attr.id))
	s.writeByte(byte(attr.Attrs))
	s.write(attr.Fg.Bytes())
	s.write(attr.Bg.Bytes())
	s.write(attr.Sp.Bytes())

	s.clearEnd = s.payload.Len()
}

func (s *Screen) writeStyle(cur *CellAttrs) {
	// Check last style sent
	if s.lastSent != cur {
		if _, ok := s.sentAttrs[cur]; !ok {
			s.sentAttrs[cur] = 0
		}

		s.writeByte(byte(OpStyle))
		s.writeEncodedInt(int(cur.id))
		s.lastSent = cur
	}
}

func (s *Screen) writeSize() {
	s.writeByte(byte(OpResize))
	s.writeEncodedInt(int(s.Size.X))
	s.writeEncodedInt(int(s.Size.Y))
}

// Flush the operations and send the final state and cursor position along with
// cell attributes that's under the cursor.  This allow the client to render the
// cursor without needing to track the cell data.
func (s *Screen) writeFlush() {
	s.flushPutOps()

	s.writeByte(byte(OpFlush))

	state := s.Mode
	if s.Busy {
		state |= ModeBusy
	}

	if s.Mouse {
		state |= ModeMouseOn
	}

	s.writeEncodedInt(int(state))
	s.writeEncodedInt(s.Cursor.X)
	s.writeEncodedInt(s.Cursor.Y)

	i := s.Cursor.Y*s.Size.X + s.Cursor.X
	if i > len(s.Buffer) {
		i = len(s.Buffer) - 1
	}
	c := s.Buffer[i]
	s.writeEncodedInt(int(c.id))
	s.writeEncodedInt(int(c.Char))
}

// Flush operations into the sink.
func (s *Screen) flush() error {
	s.writeFlush()

	data := make([]byte, s.payload.Len())
	copy(data, s.payload.Bytes())
	s.payload.Truncate(0)

	if s.sink != nil {
		// Insert the palette into the data if there are new colors that haven't
		// been sent before.
		var pn int
		for attr := range s.sentAttrs {
			if s.sentAttrs[attr] == 0 {
				s.sentAttrs[attr]++
				pn++
				s.writeEncodedInt(int(attr.id))
				s.writeByte(byte(attr.Attrs))
				s.writeEncodedInt(int(attr.Fg))
				s.writeEncodedInt(int(attr.Bg))
				s.writeEncodedInt(int(attr.Sp))
			}
		}

		if pn > 0 {
			palette := []byte{byte(OpPalette)}
			palette = append(palette, s.encodedInt(pn)...)
			palette = append(palette, s.payload.Bytes()...)
			if s.clearEnd > 0 {
				palette = append(palette, data[s.clearEnd:]...)
				data = append(data[:s.clearEnd], palette...)
			} else {
				data = append(palette, data...)
			}
			s.payload.Truncate(0)
		}

		if _, err := s.sink.Write(data); err != nil {
			s.clearEnd = 0
			return err
		}
	}

	s.clearEnd = 0

	return nil
}
