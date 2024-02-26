package ws

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

func (server *WsServer) handleNewMessage(sender *Client, event *Event) error {
	event.Payload.From = sender
	log.Println("[handle new message]", event.Type, event.Payload, event.Payload.To)
	switch event.Type {

	case SendMessageEvent:
		server.handleSendRoomMessage(sender, event)

	case JoinRoomEvent:
		server.handleJoinRoomMessage(sender, event)

	case JoinRoomPrivateEvent:
		server.handleJoinRoomPrivateMessage(sender, event)

	case LeaveRoomEvent:
		server.handleLeaveRoomMessage(sender, event)
	}

	return nil
}

func (server *WsServer) handleLeaveRoomMessage(sender *Client, event *Event) {
	roomName := event.Payload.Room.Name
	if roomName == "" {
		return
	}

	targetRoom := server.findRoomByName(roomName)
	if targetRoom == nil {
		return
	}

	delete(sender.rooms, targetRoom)

	targetRoom.unregister <- sender
}

func (server *WsServer) handleSendRoomMessage(sender *Client, event *Event) {
	roomName := event.Payload.Room.Name
	if roomName == "" {
		return
	}

	targetRoom := server.findRoomByName(roomName)
	if targetRoom == nil {
		return
	}

	if sender.isInRoom(targetRoom) {
		targetRoom.broadcast <- event
	}
}

func (server *WsServer) handleJoinRoomMessage(sender *Client, event *Event) {
	roomName := event.Payload.Room.Name
	if roomName == "" {
		return
	}

	// room := server.findRoomByName(roomName)
	// if room == nil {
	// 	room = server.createRoom(roomName, false)
	// }
	// if sender.isInRoom(room) {
	// 	return
	// }

	// sender.rooms[room] = true

	// room.register <- sender
	server.joinRoom(roomName, sender, nil)
}

func (server *WsServer) handleJoinRoomPrivateMessage(sender *Client, event *Event) {
	// target := server.findClientByID(event.Payload.To.ID.String())
	receiver := server.findClientByName(event.Payload.To.Name)
	if receiver == nil {
		return
	}

	// roomName := target.ID.String() + sender.ID.String()
	roomName := sender.Name + receiver.Name

	log.Println("[privateroom]", sender, receiver, "|", event.Payload, roomName)

	server.joinRoom(roomName, sender, receiver)
	server.joinRoom(roomName, receiver, sender)
}

func (server *WsServer) joinRoom(roomName string, sender *Client, receiver *Client) {
	room := sender.server.findRoomByName(roomName)
	if room == nil {
		room = sender.server.createRoom(roomName, receiver != nil)
	}

	if receiver == nil && room.Private {
		return
	}

	if !sender.isInRoom(room) {
		sender.rooms[room] = true
		room.register <- sender
		room.notifyClientJoined(sender)
	}
}

func (server *WsServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		select {
		case client.send <- message:
		default:
			server.unregisterClient(client)
		}
	}
}

func (server *WsServer) registerClient(client *Client) {
	server.Lock()
	defer server.Unlock()

	server.listOnlineClients(client)

	server.clients[client] = true
}

func (server *WsServer) unregisterClient(client *Client) {
	server.Lock()
	defer server.Unlock()

	if _, ok := server.clients[client]; ok {
		for room := range client.rooms {
			room.unregister <- client
		}
		server.unregister <- client
	}
}

func (server *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name, ok := r.URL.Query()["name"]
	log.Println("[ServeHTTP]", name, ok, r.URL.Query())

	if !ok || len(name[0]) < 1 {
		log.Println("url Param 'name' is missing")
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
	var found *Room
	for room := range server.rooms {
		if room.GetName() == name {
			found = room
			break
		}
	}

	return found
}

func (server *WsServer) findRoomByID(ID string) *Room {
	var found *Room
	for room := range server.rooms {
		if room.GetID() == ID {
			found = room
			break
		}
	}

	return found
}

func (server *WsServer) findClientByID(ID string) *Client {
	var found *Client
	for client := range server.clients {
		if client.ID.String() == ID {
			found = client
			break
		}
	}

	return found
}

func (server *WsServer) findClientByName(name string) *Client {
	var found *Client
	for client := range server.clients {
		if client.Name == name {
			found = client
			break
		}
	}

	return found
}

func (server *WsServer) createRoom(name string, private bool) *Room {
	room := NewRoom(name, private)
	go room.Run()
	server.rooms[room] = true

	return room
}

func (server *WsServer) listOnlineClients(client *Client) {
	message := &Event{
		Type: UserJoinedEvent,
		Payload: &ReceivedMessage{
			From: client,
		},
	}

	server.broadcast <- message.encode()
}
