package nmux

import (
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/tweekmonster/nmux/screen"
	"github.com/tweekmonster/nmux/util"

	"golang.org/x/net/websocket"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

var proc *Process

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
	ws.PayloadType = websocket.BinaryFrame

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
		var msg []byte
	loop:
		for {
			err := websocket.Message.Receive(ws, &msg)
			if err == io.EOF {
				break loop
			} else if err != nil {
				util.Print("WebSocket Err:", err)
				break loop
			} else {
				select {
				case input <- msg:
				case <-time.After(time.Second):
					// XXX: Needs investigation.  Occurs when a client reconnects and
					// sends a key event while nvim is prompting in the command line.
					// e.g. "swap recovery" prompt.
					util.Print("Input deadlock?")
					break loop
				}
			}
		}

		util.Print("Input stopped")
		closeInput()
	}()

	if err := proc.Attach(ws); err != nil {
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
	http.Handle("/nmux", websocket.Handler(websocketHandler))
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
