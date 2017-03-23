package screen

import (
	runewidth "github.com/mattn/go-runewidth"
)

const minRepeatRange = 3

// TODO: There should be a buffer that's managed elsewhere that this can
// unconditionally write to.

func (s *Screen) writeLog(message string) {
	s.payload.WriteOp(OpLog)
	s.payload.WriteStringRun(message)
}

// writeRange sends a range of characters to render.  The ranges *must* have the
// same display attributes.  Repeated characters will be sent as a "condensed"
// operation.
func (s *Screen) writeRange(i1, i2 int) {
	run := make([]rune, 0, i2-i1)
	for i := i1; i < i2; i++ {
		s.Buffer[i].Sent = true
		run = append(run, s.Buffer[i].Char)
	}

	p := s.payload
	runStart := 0
	s.writeStyle(s.Buffer[i1].CellAttrs)

	if len(run) < minRepeatRange {
		p.WriteOp(OpPut)
		p.WriteEncodedInt(i1)
		p.WriteRuneRun(run)
		return
	}

	l := len(run)
	i := 0
outer:
	for i < l-minRepeatRange {
		for j := minRepeatRange - 1; j > 0; j-- {
			if run[i] != run[i+j] {
				i++
				continue outer
			}
		}

		if i-runStart > 0 {
			p.WriteOp(OpPut)
			p.WriteEncodedInt(i1 + runStart)
			p.WriteRuneRun(run[runStart:i])
			runStart = i
		}

	scan:
		for j := i + minRepeatRange; j <= l; j++ {
			if j == l || run[i] != run[j] {
				p.WriteOp(OpPutRep)
				p.WriteEncodedInt(i1 + runStart)
				p.WriteEncodedInt(j - i)
				p.WriteEncodedInt(int(run[i]))
				runStart = j
				i = j
				break scan
			}
		}
	}

	run = run[runStart:]
	if len(run) > 0 {
		p.WriteOp(OpPut)
		p.WriteEncodedInt(i1 + runStart)
		p.WriteRuneRun(run)
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

func (s *Screen) writeTitle(title string) {
	s.payload.WriteOp(OpTitle)
	s.payload.WriteStringRun(title)
}

func (s *Screen) writeIcon(icon string) {
	s.payload.WriteOp(OpIcon)
	s.payload.WriteStringRun(icon)
}

func (s *Screen) writeBell(visual bool) {
	s.payload.WriteOp(OpBell)
	if visual {
		s.payload.WriteByte(1)
	} else {
		s.payload.WriteByte(0)
	}
}

// Flush the operations and send the final state and cursor position along with
// cell attributes that's under the cursor.  This allow the client to render the
// cursor without needing to track the cell data.
func (s *Screen) writeFlush(displayCursor bool) {
	p := s.payload
	p.WriteOp(OpFlush)

	state := s.Mode
	if s.Busy {
		state |= ModeBusy
	}

	if s.Mouse {
		state |= ModeMouseOn
	}

	if !displayCursor {
		state |= ModeRedraw
	}

	i := s.Cursor.Y*s.Size.X + s.Cursor.X
	if i >= len(s.Buffer) {
		i = len(s.Buffer) - 1
	}
	c := s.Buffer[i]

	p.WriteEncodedInts(int(state), s.Cursor.X, s.Cursor.Y, int(c.id))
	p.WriteRuneRun([]rune{c.Char})
	p.WriteEncodedInt(runewidth.RuneWidth(c.Char))
}

// Flush operations into the sink.
func (s *Screen) flush(displayCursor bool) error {
	s.writeFlush(displayCursor)
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
