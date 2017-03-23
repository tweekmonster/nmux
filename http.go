package nmux

import (
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tweekmonster/nmux/screen"
	"github.com/tweekmonster/nmux/util"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

var proc *Process

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebsocketWriter struct {
	Conn *websocket.Conn
}

func (w WebsocketWriter) Write(p []byte) (int, error) {
	err := w.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err == nil {
		return len(p), nil
	}

	return 0, err
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func websocketHandler(ws *websocket.Conn) {
	util.Print("Connection from:", ws.RemoteAddr())

	if proc == nil || !proc.IsRunning() {
		util.Debug("Starting nvim process")
		p, err := NewProcess("/tmp", 80, 20)
		if err != nil {
			util.Print("Couldn't start process:", err)
			return
		}

		proc = p
	}

	input := make(chan []byte)

	closeInput := func() {
		// Close the input without panicking.
		defer func() {
			recover()
		}()
		close(input)
	}

	go func() {
		util.Debug("Starting input loop")
	loop:
		for {
			msgType, data, err := ws.ReadMessage()
			if msgType != websocket.BinaryMessage {
				util.Print("Unsupported message type from client", ws.RemoteAddr(), msgType)
				break loop
			}

			if err == io.EOF {
				break loop
			} else if err != nil {
				util.Print("WebSocket Err:", err)
				break loop
			} else {
				select {
				case input <- data:
				case <-time.After(time.Second):
					// XXX: Needs investigation.  Occurs when a client reconnects and
					// sends a key event while nvim is prompting in the command line.
					// e.g. "swap recovery" prompt.
					util.Print("Input deadlock?")
					break loop
				}
			}
		}

		util.Print("Input stopped for client", ws.RemoteAddr())
		closeInput()
	}()

	if err := proc.Attach(WebsocketWriter{Conn: ws}); err != nil {
		util.Print("Attach err:", err)
	}

mainloop:
	for {
		select {
		case _, ok := <-proc.Deadman:
			if !ok {
				util.Print("Process ended")
				break mainloop
			}

		case data, ok := <-input:
			if !ok {
				break mainloop
			}

			op := screen.Op(data[0])

			switch op {
			case screen.OpResize:
				cols := (int(data[1]) << 8) | int(data[2])
				rows := (int(data[3]) << 8) | int(data[4])

				// XXX: This is the origin point of the deadlock mentioned above.
				if err := proc.Resize(cols, rows); err != nil {
					util.Print("Couldn't resize:", err)
					break mainloop
				}
			case screen.OpKeyboard:
				if proc != nil && proc.IsRunning() {
					if _, err := proc.Input(string(data[1:])); err != nil {
						util.Print("Input Error:", err)
						break mainloop
					}
				}
			}
		}
	}

	closeInput()

	if err := proc.Detach(); err != nil {
		util.Print("Detach err:", err)
	}

	util.Print("Connection stopped:", ws.RemoteAddr())
}

func WebServer(addr string) (io.Closer, error) {
	http.HandleFunc("/nmux", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			util.Print("Couldn't upgrade connection for", r.RemoteAddr)
			return
		}

		websocketHandler(ws)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s - %s", r.Method, r.URL.Path)
		switch r.URL.Path {
		case "/":
			r.URL.Path = "/index.html"
		case "/const.js":
			w.Write(screen.JSConst)
			return
		}

		data, err := Asset(filepath.Join("web", r.URL.Path))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	})

	server := &http.Server{
		Addr: addr,
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, err
	}

	util.Debug("Listening on", server.Addr)

	go func() {
		if err := server.Serve(tcpKeepAliveListener{listener.(*net.TCPListener)}); err != nil {
			util.Print("Server Error:", err)
		}
	}()

	return listener, nil
}
