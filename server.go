package main

import (
	"log"
	"net/http"
	"sync"
)

// WsServer to handle connections
type WsServer struct {
	sync.RWMutex
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	rooms      map[*Room]bool
}

func NewWsServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		rooms:      make(map[*Room]bool),
	}
}

// Run() iterate over all clients and get signals from clients' channels
func (server *WsServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.unregister:
			server.unregisterClient(client)
		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

func (server *WsServer) handleNewMessage(client *Client, event *Event) error {
	switch event.Type {
	case SendMessageAction:
		server.handleSendRoomMessage(client, event)
	case JoinRoomAction:
		server.handleJoinRoomMessage(client, event)
	case LeaveRoomAction:
		server.handleLeaveRoomMessage(client, event)
	}
	return nil
}

func (server *WsServer) handleLeaveRoomMessage(client *Client, event *Event) {
	roomName := event.Payload.Room
	if roomName == "" {
		return
	}

	targetRoom := server.findRoomByName(roomName)
	if targetRoom == nil {
		return
	}

	targetRoom.unregister <- client
}

func (server *WsServer) handleSendRoomMessage(client *Client, event *Event) {
	roomName := event.Payload.Room
	if roomName == "" {
		return
	}

	targetRoom := server.findRoomByName(roomName)
	if targetRoom == nil {
		targetRoom = server.createRoom(roomName)
	}

	if !client.rooms[targetRoom] {
		targetRoom.registerClientInRoom(client)
	}

	log.Println("[handleNewMessage] roomName, targetRoom", roomName, targetRoom, client, event)
	targetRoom.broadcast <- event
}

func (server *WsServer) handleJoinRoomMessage(client *Client, event *Event) {
	roomName := event.Payload.Room

	if roomName == "" {
		return
	}

	room := server.findRoomByName(roomName)
	if room == nil {
		room = server.createRoom(roomName)
	}

	client.rooms[room] = true

	room.register <- client
}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(server.clients, client)
		}
	}
}

func (server *WsServer) registerClient(client *Client) {
	server.Lock()
	defer server.Unlock()

	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	server.Lock()
	defer server.Unlock()
	// log.Println("[unregister] ", client, server.clients[client])
	if _, ok := server.clients[client]; ok {
		client.connection.Close()
		close(client.send)
		delete(server.clients, client)
	}
}

func ServeHTTP(server *WsServer, w http.ResponseWriter, r *http.Request) {
	name, ok := r.URL.Query()["name"]
	log.Println("[ServeHTTP]", name, ok, r.URL.Query())

	if !ok || len(name[0]) < 1 {
		log.Println("Url Param 'name' is missing")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, server, name[0])

	server.registerClient(client)

	log.Println("[ServeHTPP] [clients]", server.clients)

	go client.writePump()
	go client.readPump()
}

func (server *WsServer) findRoomByName(name string) *Room {
	var foundRoom *Room
	for room := range server.rooms {
		if room.GetName() == name {
			foundRoom = room
			break
		}
	}

	return foundRoom
}

func (server *WsServer) createRoom(name string) *Room {
	room := NewRoom(name)
	go room.Run()
	server.rooms[room] = true

	return room
}
