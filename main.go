// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	redis := redis.CreateRedisClient()
	db := sqlite.InitDB()
	defer db.Close()

	roomRepo := repository.NewRoomRepository(db)
	userRepo := repository.NewUserRepository(db)

	hub := ws.NewWsServer(redis, roomRepo, userRepo)
	go hub.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeHTTP(hub, w, r)
	})
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
