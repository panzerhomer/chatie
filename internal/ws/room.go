package ws

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

const welcomeMessage = "%s joined the room"

type Room struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event
	Private    bool `json:"private"`
	sync.RWMutex
}

func NewRoom(name string, private bool) *Room {
	return &Room{
		ID:         uuid.New(),
		Name:       name,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Event),
		Private:    private,
	}
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

func (room *Room) GetId() string {
	return room.ID.String()
}

func (room *Room) GetName() string {
	return room.Name
}

func (room *Room) registerClientInRoom(client *Client) {
	room.Lock()
	defer room.Unlock()

	if !room.Private {
		room.notifyClientJoined(client)
	}
	room.clients[client] = true
}

func (room *Room) unregisterClientInRoom(client *Client) {
	room.Lock()
	defer room.Unlock()

	room.notifyClientLeft(client)
	delete(room.clients, client)
}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	for client := range room.clients {
		select {
		case client.send <- message:
		default:
			delete(room.clients, client)
		}
	}
}

func (room *Room) notifyClientJoined(client *Client) {
	const welcomeMessage = "%s joined the room"

	// msg := &Event{
	// 	Type: SendMessageEvent,
	// 	Payload: &ReceivedMessage{
	// 		Room:    room,
	// 		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	// 	},
	// }

	room.broadcastToClientsInRoom([]byte(fmt.Sprintf(welcomeMessage, client.Name)))
}

func (room *Room) notifyClientLeft(client *Client) {
	const welcomeMessage = "%s left the room"

	// msg := &Event{
	// 	Type: SendMessageEvent,
	// 	Payload: &ReceivedMessage{
	// 		Room:    room,
	// 		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	// 	},
	// }

	room.broadcastToClientsInRoom([]byte(fmt.Sprintf(welcomeMessage, client.Name)))
}
