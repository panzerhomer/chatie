package main

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Room struct {
	sync.RWMutex
	ID         uuid.UUID `json:"id"`
	name       string
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event
	private    bool
}

func NewRoom(name string) *Room {
	return &Room{
		ID:         uuid.New(),
		name:       name,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Event),
		private:    false,
	}
}

func (room *Room) notifyClientJoined(client *Client) {
	const welcomeMessage = "%s joined the room"
	room.broadcastToClientsInRoom([]byte(fmt.Sprintf(welcomeMessage, client.ID)))
}

func (room *Room) notifyClientLeaved(client *Client) {
	const leaveMessage = "%s leaved the room"
	room.broadcastToClientsInRoom([]byte(fmt.Sprintf(leaveMessage, client.ID)))
}

func (room *Room) Run() {
	for {
		select {
		case client := <-room.register:
			room.registerClientInRoom(client)
		case client := <-room.unregister:
			room.unregisterClientInRoom(client)
		case message := <-room.broadcast:
			room.broadcastToClientsInRoom(message.encode())
		}
	}
}

func (room *Room) registerClientInRoom(client *Client) {
	room.Lock()
	defer room.Unlock()

	room.notifyClientJoined(client)

	room.clients[client] = true
}

func (room *Room) unregisterClientInRoom(client *Client) {
	room.Lock()
	defer room.Unlock()

	room.notifyClientLeaved(client)

	delete(room.clients, client)
}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	for client := range room.clients {
		client.send <- message
	}
}

func (room *Room) GetID() string {
	return room.ID.String()
}

func (room *Room) GetName() string {
	return room.name
}
