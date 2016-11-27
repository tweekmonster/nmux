package gui

import (
	"fmt"

	"golang.org/x/net/websocket"
)

func Connect(addr string) (*websocket.Conn, error) {
	origin := fmt.Sprintf("http://%s/", addr)
	url := fmt.Sprintf("ws://%s/nmux", addr)
	return websocket.Dial(url, "", origin)
}

func SendKey() {

}
