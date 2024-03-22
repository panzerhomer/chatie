package main

import (
	"chatie/internal/repository"
	"chatie/internal/ws"
	"chatie/pkg/db/redis"
	"chatie/pkg/db/sqlite"
	"flag"
	"log"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8080", "")

func main() {
	flag.Parse()

	redis := redis.CreateRedisClient()
	db := sqlite.InitDB()
	defer db.Close()

	roomRepo := repository.NewChatRepository(db)
	userRepo := repository.NewUserRepository(db)

	hub := ws.NewWsServer(redis, roomRepo, userRepo)
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeHTTP(hub, w, r)
	})

	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Println("server is running")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
