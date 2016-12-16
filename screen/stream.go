package screen

// TODO: There should be a buffer that's managed elsewhere that this can
// unconditionally write to.

func (s *Screen) writeLog(message string) {
	s.payload.WriteOp(OpLog)
	s.payload.WriteStringRun(message)
}

func (s *Screen) writeRange(i1, i2 int) {
	run := make([]rune, 0, i2-i1)
	p := s.payload
	runStart := i1
	repeated := 1
	run = append(run, s.Buffer[i1].Char)
	s.Buffer[i1].Sent = true

	s.writeStyle(s.Buffer[i1].CellAttrs)

	for i := i1 + 1; i < i2; i++ {
		s.Buffer[i].Sent = true

		if s.Buffer[i-1].Char != s.Buffer[i].Char {
			if len(run) > 0 && repeated > 3 {
				p.WriteOp(OpPutRep)
				p.WriteEncodedInt(runStart)
				p.WriteEncodedInt(repeated)
				p.WriteEncodedInt(int(run[0]))
				run = run[0:0]
				runStart = i
			}
			repeated = 0
			run = append(run, s.Buffer[i].Char)
			continue
		}

		if len(run) > 0 && repeated == 0 {
			p.WriteOp(OpPut)
			p.WriteEncodedInt(runStart)
			p.WriteRuneRun(run)
			run = run[0:0]
			runStart = i
		}
		repeated++
		run = append(run, s.Buffer[i].Char)
	}

	if len(run) > 0 {
		if repeated > 3 {
			p.WriteOp(OpPutRep)
			p.WriteEncodedInt(runStart)
			p.WriteEncodedInt(repeated)
			p.WriteEncodedInt(int(run[0]))
		} else {
			p.WriteOp(OpPut)
			p.WriteEncodedInt(runStart)
			p.WriteRuneRun(run)
		}
	}
}

// writeRange sends a range of characters to render.  The ranges *must* have the
// same display attributes.  A run of the same character will be sent as a
// "condensed" operation.
func (s *Screen) writeRange2(i1, i2 int) {
	run := make([]rune, 0, i2-i1)
	runs := make([][2]int, 0, i2-i1)

	firstC := s.Buffer[i1].Char
	runStart := 0
	runEnd := 0

	s.writeStyle(s.Buffer[i1].CellAttrs)
	s.nextAttrs = nil

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
	index := i1
	length := 0
	p := s.payload

	for _, r := range runs {
		if r[1]-r[0] < 4 {
			// Run is too short.
			copy(send[length:], run[r[0]:r[1]])
			length += r[1] - r[0]
		} else {
			if length > 0 {
				// Flush pending characters.
				p.WriteOp(OpPut)
				p.WriteEncodedInt(index)
				p.WriteRuneRun(send[:length])
				index += length
			}

			// Getting here means that the run is long enough to send as a condensed
			// render operation.
			p.WriteOp(OpPutRep)
			p.WriteEncodedInt(index)
			p.WriteEncodedInt(r[1] - r[0])
			p.WriteEncodedInt(int(run[r[0]]))

			index += r[1] - r[0]
			length = 0
		}
	}

	if length > 0 {
		p.WriteOp(OpPut)
		p.WriteEncodedInt(index)
		p.WriteRuneRun(send[:length])
	}

}

func (s *Screen) writeScroll(delta int) {
	p := s.payload
	p.WriteOp(OpScroll)

	p.WriteEncodedInt(int(s.DefaultAttrs.Bg))
	p.Write([]byte{byte(delta >> 8), byte(delta)})
	p.WriteEncodedInts(s.scroll.tl.Y, s.scroll.br.Y, s.scroll.tl.X, s.scroll.br.X)
}

func (s *Screen) writeClear() {
	// Clear the attribute counter of colors that are no longer used in the cells.
	for attr, count := range s.attrCounter {
		if attr.id != 0 {
			if count <= 0 && attr != s.CurAttrs {
				delete(s.attrCounter, attr)
			}

			if _, ok := s.sentAttrs[attr]; ok {
				delete(s.sentAttrs, attr)
			}
		}
	}

	s.lastSent = nil

	p := s.payload
	p.WriteOp(OpClear)

	attr := s.DefaultAttrs
	// Clear is a special case that receives the default colors before resetting
	// the palette on the client side.
	p.WriteEncodedInt(int(attr.id))
	p.WriteByte(byte(attr.Attrs))
	p.WriteEncodedInts(int(attr.Fg), int(attr.Bg), int(attr.Sp))

	s.clearEnd = p.Len()
}

func (s *Screen) writeStyle(cur *CellAttrs) {
	// Check last style sent
	if s.lastSent != cur {
		if _, ok := s.sentAttrs[cur]; !ok {
			s.sentAttrs[cur] = 0
		}

		s.payload.WriteOp(OpStyle)
		s.payload.WriteEncodedInt(int(cur.id))
		s.lastSent = cur
	}
}

func (s *Screen) writeSize() {
	s.payload.WriteOp(OpResize)
	s.payload.WriteEncodedInts(s.Size.X, s.Size.Y)
}

// Flush the operations and send the final state and cursor position along with
// cell attributes that's under the cursor.  This allow the client to render the
// cursor without needing to track the cell data.
func (s *Screen) writeFlush() {
	p := s.payload
	p.WriteOp(OpFlush)

	state := s.Mode
	if s.Busy {
		state |= ModeBusy
	}

	if s.Mouse {
		state |= ModeMouseOn
	}

	i := s.Cursor.Y*s.Size.X + s.Cursor.X
	if i >= len(s.Buffer) {
		i = len(s.Buffer) - 1
	}
	c := s.Buffer[i]

	p.WriteEncodedInts(int(state), s.Cursor.X, s.Cursor.Y, int(c.id), int(c.Char))
}

// Flush operations into the sink.
func (s *Screen) flush() error {
	s.writeFlush()
	defer s.payload.Truncate(0)

	data := s.payload.Bytes()

	if s.sink != nil {
		// Insert the palette into the data if there are new colors that haven't
		// been sent before.
		var pn int
		var attrs []*CellAttrs
		var cid int
		var colors []Color
		cmap := map[Color]int{}

		defer s.buf.Truncate(0)

		for attr := range s.sentAttrs {
			if s.sentAttrs[attr] == 0 {
				s.sentAttrs[attr]++
				attrs = append(attrs, attr)

				if _, ok := cmap[attr.Fg]; !ok {
					colors = append(colors, attr.Fg)
					cmap[attr.Fg] = cid
					cid++
				}

				if _, ok := cmap[attr.Bg]; !ok {
					colors = append(colors, attr.Bg)
					cmap[attr.Bg] = cid
					cid++
				}

				if _, ok := cmap[attr.Sp]; !ok {
					colors = append(colors, attr.Sp)
					cmap[attr.Sp] = cid
					cid++
				}

				pn++
			}
		}

		if pn > 0 {
			s.buf.WriteEncodedInt(cid)

			for _, color := range colors {
				s.buf.WriteColor(color)
			}

			s.buf.WriteEncodedInt(len(attrs))

			for _, attr := range attrs {
				s.buf.WriteEncodedInt(int(attr.id))
				s.buf.WriteByte(byte(attr.Attrs))
				s.buf.WriteEncodedInt(cmap[attr.Fg])
				s.buf.WriteEncodedInt(cmap[attr.Bg])
				s.buf.WriteEncodedInt(cmap[attr.Sp])
			}

			offset := 1 + s.buf.Len()
			palette := make([]byte, offset+len(data))
			ps := palette[:s.clearEnd]
			ps = append(ps, byte(OpPalette))
			ps = append(ps, s.buf.Bytes()...)

			copy(palette[:s.clearEnd+offset], data[:s.clearEnd])
			copy(palette[s.clearEnd+offset:], data[s.clearEnd:])
			data = palette
		}

		if _, err := s.sink.Write(data); err != nil {
			s.clearEnd = 0
			return err
		}
	}

	s.clearEnd = 0

	return nil
}
