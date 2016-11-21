package screen

import "log"

func (s *Screen) redrawOp(op string, args *opArgs) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Println("PANIC:", r)
	// 	}
	// }()

	switch op {
	case "resize":
		s.flushPutOps()
		s.setSize(args.Int(), args.Int())
		s.writeSize()

	case "clear":
		s.flushPutOps()
		s.charUpdate.X = 0
		s.charUpdate.Y = -1
		s.charTracking = false
		i1 := 0
		i2 := len(s.Buffer)
		for i1 < i2 {
			s.setChar(i1, ' ')
			i1++
		}
		s.charTracking = true
		s.Cursor.X = 0
		s.Cursor.Y = 0

		s.writeClear()

	case "eol_clear":
		s.clearLine(s.Cursor.X, s.Cursor.Y)

	case "cursor_goto":
		s.flushPutOps()
		y := args.Int()
		x := args.Int()
		s.setCursor(x, y)

	case "update_fg":
		s.DefaultAttrs.Fg = Color(args.Int())

	case "update_bg":
		s.DefaultAttrs.Bg = Color(args.Int())

	case "update_sp":
		s.DefaultAttrs.Sp = Color(args.Int())

	case "highlight_set":
		s.flushPutOps()

		m := args.Map()
		attrs := s.DefaultAttrs

		if c, ok := m.Int64("foreground"); ok {
			attrs.Fg = Color(c)
		}

		if c, ok := m.Int64("background"); ok {
			attrs.Bg = Color(c)
		}

		if c, ok := m.Int64("special"); ok {
			attrs.Sp = Color(c)
		}

		if b, ok := m.Bool("reverse"); ok && b {
			attrs.Attrs |= AttrReverse
		}

		if b, ok := m.Bool("italic"); ok && b {
			attrs.Attrs |= AttrItalic
		}

		if b, ok := m.Bool("bold"); ok && b {
			attrs.Attrs |= AttrBold
		}

		if b, ok := m.Bool("underline"); ok && b {
			attrs.Attrs |= AttrUnderline
		}

		if b, ok := m.Bool("undercurl"); ok && b {
			attrs.Attrs |= AttrUndercurl
		}

		s.CurAttrs = attrs
		s.writeStyle(attrs)

	case "put":
		i := s.Cursor.Y*s.Size.X + s.Cursor.X
		for _, c := range args.String() {
			s.setChar(i, c)
			i++
		}

	case "set_scroll_region":
		s.scroll.tl.Y = args.Int()
		s.scroll.br.Y = args.Int()
		s.scroll.tl.X = args.Int()
		s.scroll.br.X = args.Int()

	case "scroll":
		s.flushPutOps()

		amount := args.Int()
		sr := s.scroll
		blank := make([]Cell, (sr.br.X-sr.tl.X)+1)
		for i := range blank {
			blank[i].Char = ' '
			blank[i].CellAttrs = s.CurAttrs
		}

		ys := amount
		h := (sr.br.Y - sr.tl.Y) + 1
		if amount < 0 {
			// Down
			ys = -amount
		}

		var sy, dy int

		// Copying must go from top to bottom regardless of the scroll direction.
		for y := ys; y < h; y++ {
			if amount < 0 {
				dy = sr.br.Y + ys - y
				sy = dy + amount
			} else {
				sy = sr.tl.Y + y
				dy = sy - amount
			}

			sy *= s.Size.X
			dy *= s.Size.X

			src := s.Buffer[sy+sr.tl.X : sy+sr.br.X+1]
			dst := s.Buffer[dy+sr.tl.X : dy+sr.br.X+1]

			copy(dst, src)
			copy(src, blank) // Always blank the source line.
		}

		s.writeScroll(amount)

	case "set_title":
		s.Title = args.String()

	case "set_icon":

	case "mouse_on":
		s.Mouse = true
		log.Println("Mouse Enabled")

	case "mouse_off":
		s.Mouse = false
		log.Println("Mouse Disabled")

	case "busy_start":
		fallthrough
	case "busy_on":
		s.Busy = true
		log.Println("Busy")

	case "busy_stop":
		fallthrough
	case "busy_off":
		s.Busy = false
		log.Println("Not Busy")

	case "suspend":

	case "bell":
		log.Println("Bell")

	case "visual_bell":
		log.Println("Visual bell")

	case "mode_change":
		mode := args.String()
		log.Println("Mode change:", mode)

		switch mode {
		case "normal":
			s.Mode = ModeNormal
		case "insert":
			s.Mode = ModeInsert
		case "replace":
			s.Mode = ModeReplace
		}

	case "popupmenu_show":
		// TODO

	case "popupmenu_select":

	case "popupmenu_hide":

	default:
		log.Println("Unknown redraw op:", op)
	}
}

func (s *Screen) RedrawHandler(updates ...[]interface{}) {
	if len(updates) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

oploop:
	for _, args := range updates {
		var op string
		switch n := args[0].(type) {
		case string:
			op = n
		default:
			log.Println("Unknown Op:", n)
			break oploop
		}

		for _, u := range args[1:] {
			switch a := u.(type) {
			case []interface{}:
				s.redrawOp(op, &opArgs{args: a})

			default:
				log.Printf("Unknown arguments for op '%s': %#v\n", op, a)
			}

		}
	}

	if err := s.flush(); err != nil {
		log.Println("Couldn't flush data:", err)
	}

	// s.dump()
}
