package ws

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Room struct {
	sync.RWMutex
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Private    bool      `json:"private"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event
}

func NewRoom(name string, private bool) *Room {
	return &Room{
		ID:         uuid.New(),
		Name:       name,
		Private:    private,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Event),
	}
}

func (room *Room) notifyClientJoined(client *Client) {
	const welcomeMessage = "%s joined the room\n"
	room.broadcastToClientsInRoom([]byte(fmt.Sprintf(welcomeMessage, client.ID)))
}

func (room *Room) notifyClientLeaved(client *Client) {
	const leaveMessage = "%s leaved the room\n"
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

	if !room.Private {
		room.notifyClientJoined(client)
	}
	// room.notifyClientJoined(client)

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
	return room.Name
}
