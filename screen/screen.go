package screen

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"runtime/debug"
	"sync"
)

type CellAttrs struct {
	Attrs Attr
	Fg    Color
	Bg    Color
	Sp    Color
}

type Cell struct {
	Char rune
	CellAttrs
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
	mu           sync.Mutex
	Size         Vector2 // Screen size.
	Cursor       Vector2 // Cursor postion.
	Title        string
	Mode         Mode         // Current mode.
	Mouse        bool         // Mouse state. This updates Mode.
	Busy         bool         // Busy state. This updates Mode.
	DefaultAttrs CellAttrs    // Default attributes for updating attributes.
	CurAttrs     CellAttrs    // Current attributes for new characters.
	scroll       ScrollRegion // Region to scroll.
	Buffer       []Cell       // The rendered screen.
	charUpdate   Vector2      // Tracks ranges of characters being set.
	charTracking bool         // Whether or not to track characters being set.
	payload      bytes.Buffer
	writer       *bufio.Writer
	sink         io.Writer
}

// NewScreen creates a new screen.
func NewScreen(w, h int) *Screen {
	s := &Screen{
		charUpdate: Vector2{0, -1},
		Mode:       ModeNormal | ModeMouseOn,
	}

	s.writer = bufio.NewWriter(&s.payload)

	s.setSize(w, h)
	return s
}

// SetSink sets the writer to receive operation writes.  The full screen is sent
// to the new client.
func (s *Screen) SetSink(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.flush()
	s.sink = w

	s.writeSize()
	s.writeStyle(s.DefaultAttrs)
	s.writeClear()

	lastIndex := 0
	lastCell := s.Buffer[0]

	for i, c := range s.Buffer {
		if c.CellAttrs != lastCell.CellAttrs {
			s.writeStyle(lastCell.CellAttrs)
			s.writeRange(lastIndex, i)
			lastIndex = i
			lastCell = c
		}
	}

	if lastIndex < len(s.Buffer) {
		s.writeStyle(lastCell.CellAttrs)
		s.writeRange(lastIndex, len(s.Buffer))
	}

	s.flush()
}

func (s *Screen) flushPutOps() {
	if s.charUpdate.X >= 0 && s.charUpdate.Y >= 0 && s.charUpdate.Y-s.charUpdate.X > 0 {
		s.writeRange(s.charUpdate.X, s.charUpdate.Y)
		s.charUpdate.X = s.charUpdate.Y
		s.charUpdate.Y = -1
	}
}

// clearLine is a helper for clearing a line.
func (s *Screen) clearLine(x, y int) {
	s.flushPutOps()

	i1 := y*s.Size.X + x
	i2 := i1 + (s.Size.X - x)

	for i := i1; i < i2; i++ {
		s.Buffer[i].Char = ' '
		s.Buffer[i].CellAttrs = s.CurAttrs
	}

	s.writeRange(i1, i2)
}

func (s *Screen) setCursor(x, y int) {
	s.Cursor.X = x
	s.Cursor.Y = y
	s.charUpdate.X = y*s.Size.X + x
	s.charUpdate.Y = -1
}

// setChar sets a Cell's character, attributes, and colors.  It also updates the
// cursor position and the range of cells to be included when flushing put
// operations.
func (s *Screen) setChar(index int, c rune) {
	s.Buffer[index] = Cell{
		Char:      c,
		CellAttrs: s.CurAttrs,
	}

	index++
	if s.charTracking {
		s.charUpdate.Y = index
	}
	s.Cursor.X = index % s.Size.X
	s.Cursor.Y = index / s.Size.X
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

		for i := curSize; i < newSize; i++ {
			s.setChar(i, ' ')
		}
	}

	s.charUpdate.X = 0
	s.charUpdate.Y = -1
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
