package screen

import "encoding/binary"

// TODO: There should be a buffer that's managed elsewhere that this can
// unconditionally write to.

func (s *Screen) writeBinary(v interface{}) error {
	return binary.Write(s.writer, binary.BigEndian, v)
}

func (s *Screen) writeByte(c byte) error {
	return s.writer.WriteByte(c)
}

func (s *Screen) writeRune(r rune) (int, error) {
	return s.writer.WriteRune(r)
}

func (s *Screen) write(p []byte) (int, error) {
	return s.writer.Write(p)
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
				s.writeBinary(OpPut)
				s.writeBinary(uint16(i))
				s.writeBinary(uint16(j))
				i += j

				for _, c := range send[:j] {
					s.writeBinary(uint16(c))
				}
			}

			// Getting here means that the run is long enough to send as a condensed
			// render operation.
			s.writeBinary(OpPutRep)
			s.writeBinary(uint16(i))
			s.writeBinary(uint16(r[1] - r[0]))
			s.writeBinary(uint16(run[r[0]]))

			i += r[1] - r[0]
			j = 0
		}
	}

	if j > 0 {
		s.writeBinary(OpPut)
		s.writeBinary(uint16(i))
		s.writeBinary(uint16(j))
		for _, c := range send[:j] {
			s.writeBinary(uint16(c))
		}
	}
}

func (s *Screen) writeScroll(delta int) {
	s.writeBinary(OpScroll)
	s.write(s.DefaultAttrs.Bg.Bytes())
	s.writeBinary(int16(delta))
	s.writeBinary(uint16(s.scroll.tl.Y))
	s.writeBinary(uint16(s.scroll.br.Y))
	s.writeBinary(uint16(s.scroll.tl.X))
	s.writeBinary(uint16(s.scroll.br.X))
}

func (s *Screen) writeClear() {
	s.writeBinary(OpClear)
}

func (s *Screen) writeStyle(cur CellAttrs) {
	s.writeBinary(OpStyle)
	s.writeBinary(cur.Attrs)
	s.write(cur.Fg.Bytes())
	s.write(cur.Bg.Bytes())
	s.write(cur.Sp.Bytes())
}

func (s *Screen) writeSize() {
	s.writeBinary(OpResize)
	s.writeBinary(uint16(s.Size.X))
	s.writeBinary(uint16(s.Size.Y))
}

// Flush the operations and send the final state and cursor position along with
// cell attributes that's under the cursor.  This allow the client to render the
// cursor without needing to track the cell data.
func (s *Screen) writeFlush() {
	s.flushPutOps()

	s.writeBinary(OpFlush)

	state := s.Mode
	if s.Busy {
		state |= ModeBusy
	}

	if s.Mouse {
		state |= ModeMouseOn
	}

	s.writeBinary(state)
	s.writeBinary(uint16(s.Cursor.X))
	s.writeBinary(uint16(s.Cursor.Y))

	i := s.Cursor.Y*s.Size.X + s.Cursor.X
	if i > len(s.Buffer) {
		i = len(s.Buffer) - 1
	}
	c := s.Buffer[i]
	s.writeBinary(uint16(c.Char))
	s.writeBinary(c.Attrs)
	s.write(c.Fg.Bytes())
	s.write(c.Bg.Bytes())
	s.write(c.Sp.Bytes())
}

// Flush operations into the sink.
func (s *Screen) flush() error {
	s.writeFlush()
	s.writer.Flush()

	if s.sink != nil {
		if _, err := s.sink.Write(s.payload.Bytes()); err != nil {
			return err
		}
	}

	s.payload.Truncate(0)
	return nil
}
