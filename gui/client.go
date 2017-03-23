package gui

import (
	"fmt"
	"io"
	"time"

	screen "github.com/tweekmonster/nmux/screen"
	"github.com/tweekmonster/nmux/util"

	"github.com/gorilla/websocket"
)

type Client struct {
	Addr       string
	Conn       *websocket.Conn
	firstRun   bool
	palette    map[int]screen.CellAttrs
	ClearBG    screen.Color
	curStyle   screen.CellAttrs
	resizeChan chan screen.Vector2
	grid       screen.Vector2
	bufferSize screen.Vector2
	app        *App
}

func NewClient(addr string, app *App) (*Client, error) {
	c := &Client{
		Addr:       addr,
		app:        app,
		firstRun:   true,
		palette:    make(map[int]screen.CellAttrs),
		resizeChan: make(chan screen.Vector2),
	}

	if err := c.Connect(); err != nil {
		return nil, err
	}

	go c.handleIncoming()
	go c.resizeDebounce()

	return c, nil
}

func (c *Client) Connect() error {
	url := fmt.Sprintf("ws://%s/nmux", c.Addr)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	c.Conn = conn
	return nil
}

func (c *Client) resizeDebounce() {
	var size screen.Vector2
	var ok bool

	size, ok = <-c.resizeChan
	if !ok {
		return
	}

loop:
	for {
		select {
		case size, ok = <-c.resizeChan:
			if !ok {
				break loop
			}
		case <-time.After(time.Millisecond * 100):
			var payload [5]byte

			payload[0] = uint8(screen.OpResize)
			payload[1] = uint8(size.X >> 8)
			payload[2] = uint8(size.X & 0xff)
			payload[3] = uint8(size.Y >> 8)
			payload[4] = uint8(size.Y & 0xff)

			c.grid = size
			if err := c.Conn.WriteMessage(websocket.BinaryMessage, payload[:]); err != nil {
				util.Debug("Client write error:", err)
				break loop
			}

			size, ok = <-c.resizeChan
			if !ok {
				break loop
			}
		}
	}
}

func (c *Client) Resize(w, h int) {
	if c.grid.X == w && c.grid.Y == h {
		return
	}
	c.resizeChan <- screen.Vector2{X: w, Y: h}
}

func (c *Client) parseOps(r *StreamReader) {
	win := app.GetWindow(0)

	if win == nil {
		win, _ = app.NewWindow(80, 20)
	}

	for r.Remaining() > 0 {
		op := r.ReadOp()

		if c.firstRun && op != screen.OpResize {
			return
		}

		switch op {
		case screen.OpResize:
			c.firstRun = false
			w := r.ReadEint32()
			h := r.ReadEint32()

			c.grid.X = w
			c.grid.Y = h
			win.SetGrid(w, h)

		case screen.OpPalette:
			colorsLen := r.ReadEint32()
			colors := make([]int, 0, colorsLen)

			for colorsLen > 0 {
				colors = append(colors, r.ReadEint32())
				colorsLen--
			}

			paletteLen := r.ReadEint32()
			for paletteLen > 0 {
				id := r.ReadEint32()
				attrs := r.ReadUint8()
				fg := colors[r.ReadEint32()]
				bg := colors[r.ReadEint32()]
				sp := colors[r.ReadEint32()]
				c.palette[id] = screen.CellAttrs{
					Attrs: screen.Attr(attrs),
					Fg:    screen.Color(fg),
					Bg:    screen.Color(bg),
					Sp:    screen.Color(sp),
				}
				paletteLen--
			}

		case screen.OpStyle:
			id := r.ReadEint32()
			if style, ok := c.palette[id]; ok {
				c.curStyle = style
			} else {
				util.Debug("Unknown style ID:", id)
			}

		case screen.OpPut:
			index := r.ReadEint32()
			str := r.ReadString()

			win.PutString(str, index, c.curStyle)

		case screen.OpPutRep:
			index := r.ReadEint32()
			length := r.ReadEint32()
			char := rune(r.ReadEint32())

			win.PutRepeatedString(char, length, index, c.curStyle)

		case screen.OpScroll:
			bg := r.ReadEint32()
			delta := r.ReadInt16()
			top := r.ReadEint32()
			bottom := r.ReadEint32()
			left := r.ReadEint32()
			right := r.ReadEint32()

			win.Scroll(int(delta), top, bottom, left, right, screen.Color(bg))

		case screen.OpClear:
			id := r.ReadEint32()
			attrs := r.ReadUint8()
			fg := r.ReadEint32()
			bg := r.ReadEint32()
			sp := r.ReadEint32()

			c.palette = make(map[int]screen.CellAttrs)
			p := screen.CellAttrs{
				Attrs: screen.Attr(attrs),
				Fg:    screen.Color(fg),
				Bg:    screen.Color(bg),
				Sp:    screen.Color(sp),
			}

			c.palette[id] = p
			c.ClearBG = p.Bg

			win.Clear(p.Bg)

		case screen.OpFlush:
			mode := r.ReadEint32()
			cursorX := r.ReadEint32()
			cursorY := r.ReadEint32()
			id := r.ReadEint32()
			char := r.ReadString()
			width := r.ReadEint32()

			win.Flush(mode, char, width, screen.Vector2{X: cursorX, Y: cursorY}, c.palette[id])

		case screen.OpLog:
			util.Print("[Server Log]", r.ReadString())

		default:
			util.Debug("Unknown Op:", op)
		}
	}
}

func (c *Client) handleIncoming() {
loop:
	for {
		msgType, data, err := c.Conn.ReadMessage()
		if msgType != websocket.BinaryMessage {
			util.Print("Unsupported message type from server: %d", msgType)
			break loop
		}

		if err == io.EOF {
			break loop
		} else if err != nil {
			util.Debug("Error reading message from server:", err)
		} else {
			// util.Print("Data:", data)
			r := StreamReader{Data: data}
			c.parseOps(&r)
		}
	}
}

func (c *Client) SendInput(input string) {
	msg := append([]byte{byte(screen.OpKeyboard)}, []byte(input)...)
	c.Conn.WriteMessage(websocket.BinaryMessage, msg)
}
