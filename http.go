package nmux

import (
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/tweekmonster/nmux/screen"

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
	log.Println("Connection from:", ws.RemoteAddr())

	input := make(chan []byte)
	ws.PayloadType = websocket.BinaryFrame

	go func() {
		var msg []byte
	loop:
		for {
			err := websocket.Message.Receive(ws, &msg)
			if err == io.EOF {
				break loop
			} else if err != nil {
				log.Println("WebSocket Err:", err)
				break loop
			} else {
				select {
				case input <- msg:
				case <-time.After(time.Second):
					// XXX: Needs investigation.  Occurs when a client reconnects and
					// sends a key event while nvim is prompting in the command line.
					// e.g. "swap recovery" prompt.
					log.Println("Input deadlock?")
					break loop
				}
			}
		}

		log.Println("Input stopped")
		close(input)
	}()

	if proc != nil && proc.IsRunning() {
		proc.SetSink(ws)
	}

	for data := range input {
		op := screen.Op(data[0])

		switch op {
		case screen.OpResize:
			cols := (int(data[1]) << 8) | int(data[2])
			rows := (int(data[3]) << 8) | int(data[4])

			if proc != nil && proc.IsRunning() {
				// XXX: This is the origin point of the deadlock mentioned above.
				proc.Resize(cols, rows)
			} else {
				p, err := NewProcess("", cols, rows)
				if err != nil {
					log.Println("Proc Err:", err)
				} else {
					proc = p
					proc.SetSink(ws)
				}
			}

		case screen.OpKeyboard:
			if proc != nil && proc.IsRunning() {
				proc.Input(string(data[1:]))
			}

		}
	}

	select {
	case _, ok := <-input:
		if ok {
			close(input)
		}
	}

	log.Println("Connection stopped:", ws.RemoteAddr())
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

	go func() {
		if err := server.Serve(tcpKeepAliveListener{listener.(*net.TCPListener)}); err != nil {
			log.Println("Server Error:", err)
		}
	}()

	return listener, nil
}
