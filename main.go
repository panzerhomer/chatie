package main

import (
	"chatie/ws"
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http server address")

func main() {
	flag.Parse()

	wsServer := ws.NewWsServer()
	go wsServer.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsServer.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}
