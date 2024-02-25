package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http server address")

func main() {
	flag.Parse()

	// http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
	// 	ServeWs(w, r)
	// })

	wsServer := NewWsServer()
	go wsServer.Run()

	// http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeHTTP(wsServer, w, r)
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}
