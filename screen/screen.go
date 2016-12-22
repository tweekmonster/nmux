package screen

import (
	"io"
	"log"
	"runtime/debug"
	"sync"

	runewidth "github.com/mattn/go-runewidth"
)

type CellAttrs struct {
	id    uint32
	Attrs Attr
	Fg    Color
	Bg    Color
	Sp    Color
}

type Cell struct {
	Char rune
	Sent bool
	*CellAttrs
}

type Vector2 struct {
	X int
	Y int
}

type ScrollRegion struct {
	tl Vector2
	br Vector2
}

type Screen struct {
	mu              sync.Mutex
	Size            Vector2 // Screen size.
	Cursor          Vector2 // Cursor postion.
	lastCursorIndex int     // Last position of the set_cursor command.
	Title           string

	// Current mode.
	Mode Mode

	// Mouse state. This updates Mode.
	Mouse bool

	// Busy state. This updates Mode.
	Busy bool

	// Default attributes for updating attributes.
	DefaultAttrs *CellAttrs

	// Current attributes for new characters.
	CurAttrs *CellAttrs

	// The next CellAttrs to write out.
	nextAttrs *CellAttrs

	// To avoid assigning a new ID to previously seen attributes.  Cleared after
	// nvim sends a clear command.
	attrCounter map[*CellAttrs]int

	// Keeps track of what attributes have been sent to the client.  Cleared after
	// nvim sends a clear command.
	sentAttrs map[*CellAttrs]int
	lastSent  *CellAttrs

	attrID uint32

	flushCount int

	// Region to scroll.
	scroll ScrollRegion

	// The rendered screen.
	Buffer []Cell

	// The position of a clear operation in the output buffer.  This allows for
	// the palette to be injected right after.
	clearEnd int

	payload *StreamBuffer
	buf     *StreamBuffer // Reusable buffer for writing palette data.
	sink    io.Writer
}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	attrs := &CellAttrs{}

	s := &Screen{
		lastCursorIndex: -1,
		DefaultAttrs:    attrs,
		CurAttrs:        attrs,
		attrCounter:     make(map[*CellAttrs]int),
		sentAttrs:       make(map[*CellAttrs]int),
		Mode:            ModeNormal | ModeMouseOn,
		payload:         &StreamBuffer{},
		buf:             &StreamBuffer{},
	}

	s.setSize(w, h)
	return s
}

// SetSink sets the writer to receive operation writes.  The full screen is sent
// to the new client.
func (s *Screen) SetSink(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// XXX: Need to investigate the command line prompt blocking screen updates.
	s.sink = w
	s.writeSize()
	s.writeClear()
	s.flushScreen(true)
	s.flush(true)
}

func (s *Screen) flushScreen(all bool) {
	prev := -1
	var attrs *CellAttrs

	for i, c := range s.Buffer {
		if all || !c.Sent {
			if prev == -1 {
				prev = i
				attrs = c.CellAttrs
			} else if attrs != c.CellAttrs {
				s.writeRange(prev, i)
				prev = i
				attrs = c.CellAttrs
			}
		} else {
			if prev != -1 {
				s.writeRange(prev, i)
				prev = -1
				attrs = nil
			}
		}
	}

	if prev != -1 {
		s.writeRange(prev, len(s.Buffer))
	}
}

func (s *Screen) setCellAttrs(index int, attrs *CellAttrs) {
	cellAttr := s.Buffer[index].CellAttrs
	if cellAttr != attrs {
		if cellAttr != nil {
			s.attrCounter[cellAttr]--
		}
		s.attrCounter[attrs]++
		s.Buffer[index].CellAttrs = attrs
	}
}

// clearLine is a helper for clearing a line.
func (s *Screen) clearLine(x, y int) {
	i1 := y*s.Size.X + x
	i2 := i1 + (s.Size.X - x)

	for i := i1; i < i2; i++ {
		s.Buffer[i].Char = ' '
		s.Buffer[i].Sent = false
		s.setCellAttrs(i, s.DefaultAttrs)
	}
}

func (s *Screen) setCursor(x, y int) {
	s.Cursor.X = x
	s.Cursor.Y = y

	index := y*s.Size.X + x
	s.lastCursorIndex = index
}

// setChar sets a Cell's character, attributes, and colors.  It also updates the
// cursor position and the range of cells to be included when flushing put
// operations.
func (s *Screen) setChar(index int, c rune) {
	s.Buffer[index].Char = c
	s.Buffer[index].Sent = false
	s.setCellAttrs(index, s.CurAttrs)

	var w = runewidth.RuneWidth(c)
	if w == 2 {
		s.Buffer[index+1].Char = ' '
		s.Buffer[index+1].Sent = false
		s.setCellAttrs(index+1, s.CurAttrs)
	}
	index += w
	s.Cursor.X = index % s.Size.X
	s.Cursor.Y = index / s.Size.X
}

func (s *Screen) clearScreen() {
	for i := 0; i < len(s.Buffer); i++ {
		s.setChar(i, ' ')
	}
}

// setSize sets the Buffer's size.
func (s *Screen) setSize(w, h int) {
	if w == 0 || h == 0 {
		log.Printf("%s", debug.Stack())
		return
	}

	curSize := s.Size.X * s.Size.Y
	newSize := w * h

	s.Size.X = w
	s.Size.Y = h

	if newSize > curSize {
		buf := make([]Cell, newSize)
		copy(buf, s.Buffer)
		s.Buffer = buf
		// for i := curSize; i < newSize; i++ {
		// 	s.setChar(i, ' ')
		// }
	}

	// Reset the scroll region on resize.
	s.scroll.tl.X = 0
	s.scroll.tl.Y = 0
	s.scroll.br.X = s.Size.X - 1
	s.scroll.br.Y = s.Size.Y - 1

	s.clearScreen()
}

func (s *Screen) dump() {
	var line string
	var i int
	var l = len(s.Buffer)

	for i < l {
		line += string(s.Buffer[i].Char)
		i++
		if i%s.Size.X == 0 {
			log.Println(line)
			line = ""
		}
	}
}
