package ws

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

const welcomeMessage = "%s joined the chat"

type Chat struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Private    bool      `json:"private"`
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Event
	sync.RWMutex
}

func NewChat(name string, private bool) *Chat {
	return &Chat{
		ID:         uuid.New(),
		Name:       name,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Event),
		Private:    private,
	}
}

func (c *Chat) Run() {
	for {
		select {
		case client := <-c.register:
			c.registerClientInRoom(client)
		case client := <-c.unregister:
			c.unregisterClientInRoom(client)
		case message := <-c.broadcast:
			c.broadcastToClientsInRoom(message.encode())
		}
	}
}

func (c *Chat) GetId() string {
	return c.ID.String()
}

func (c *Chat) GetName() string {
	return c.Name
}

func (c *Chat) registerClientInRoom(client *Client) {
	c.Lock()
	defer c.Unlock()

	if !c.Private {
		c.notifyClientJoined(client)
	}
	c.clients[client] = true
}

func (c *Chat) unregisterClientInRoom(client *Client) {
	c.Lock()
	defer c.Unlock()

	c.notifyClientLeft(client)
	delete(c.clients, client)
}

func (c *Chat) broadcastToClientsInRoom(message []byte) {
	for client := range c.clients {
		select {
		case client.send <- message:
		default:
			delete(c.clients, client)
		}
	}
}

func (c *Chat) notifyClientJoined(client *Client) {
	const welcomeMessage = "%s joined the c"

	// msg := &Event{
	// 	Type: SendMessageEvent,
	// 	Payload: &ReceivedMessage{
	// 		Room:    c,
	// 		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	// 	},
	// }

	c.broadcastToClientsInRoom([]byte(fmt.Sprintf(welcomeMessage, client.Name)))
}

func (c *Chat) notifyClientLeft(client *Client) {
	const welcomeMessage = "%s left the c"

	// msg := &Event{
	// 	Type: SendMessageEvent,
	// 	Payload: &ReceivedMessage{
	// 		Room:    c,
	// 		Message: fmt.Sprintf(welcomeMessage, client.GetName()),
	// 	},
	// }

	c.broadcastToClientsInRoom([]byte(fmt.Sprintf(welcomeMessage, client.Name)))
}
