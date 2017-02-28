package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tweekmonster/nmux"
	"github.com/tweekmonster/nmux/gui"
)

func main() {
	server := flag.Bool("server", false, "Run as server")
	addr := flag.String("addr", ":9999", "addr:port to listen on")

	flag.Parse()

	if !*server {
		gui.Main(*addr)
		return
	}

	if addr == nil {
		flag.PrintDefaults()
		return
	}

	listener, err := nmux.WebServer(*addr)
	if err != nil {
		log.Println("Error:", err)
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	if err := listener.Close(); err != nil {
		log.Println("Error:", err)
	}
}
