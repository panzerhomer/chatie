package ws

import (
	"chatie/internal/models"
	"fmt"
	"log"

	"github.com/google/uuid"
)

const welcomeMessage = "%s joined the room"
const leaveMessage = "%s left the room"

type WsChat struct {
	id         uuid.UUID
	name       string
	chatID     int
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *WebsocketMessage
	private    bool
	wsServer   *WsServer
}

func NewChat(wsServer *WsServer, name string, private bool) *WsChat {
	return &WsChat{
		id:         uuid.New(),
		name:       name,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *WebsocketMessage),
		private:    private,
		wsServer:   wsServer,
	}
}

func (c *WsChat) Run() {
	go c.subscribeToChatMessages()

	for {
		select {
		case client := <-c.register:
			c.registerClientInChat(client)
		case client := <-c.unregister:
			c.unregisterClientInChat(client)
		case message := <-c.broadcast:
			c.publishChatMessage(message.encode())
		}
	}
}

func (c *WsChat) publishChatMessage(message []byte) {
	err := c.wsServer.redis.Publish(ctx, c.GetName(), message).Err()
	if err != nil {
		log.Println(err)
	}
}

func (c *WsChat) subscribeToChatMessages() {
	pubsub := c.wsServer.redis.Subscribe(ctx, c.GetName())

	ch := pubsub.Channel()

	for msg := range ch {
		c.broadcastToChatClients([]byte(msg.Payload))
	}
}

func (c *WsChat) registerClientInChat(client *Client) {
	// fmt.Println("[char]", c.GetName(), " clients join", c.clients)
	if !c.IsPrivate() {
		c.notifyClientJoined(client)
		// c.addUserToOnlineSet(client)
	}
	c.clients[client] = true
}

func (c *WsChat) unregisterClientInChat(client *Client) {
	if _, ok := c.clients[client]; ok {
		delete(c.clients, client)
		c.notifyClientLeft(client)
		// c.removeUserFromOnlineSet(client)
		// fmt.Println("[char]", c.GetName(), " clients left", c.clients)
	}
}

func (c *WsChat) broadcastToChatClients(message []byte) {
	for client := range c.clients {
		client.send <- message
	}
}

func (c *WsChat) notifyClientJoined(client *Client) {
	message := &WebsocketMessage{
		Action: UserJoinedAction,
		Target: c.GetID(),
		Message: &models.Message{
			Text: fmt.Sprintf(welcomeMessage, client.GetName()),
		},
	}

	c.publishChatMessage(message.encode())
}

func (c *WsChat) notifyClientLeft(client *Client) {
	message := &WebsocketMessage{
		Action: UserLeftAction,
		Target: c.GetID(),
		Message: &models.Message{
			Text: fmt.Sprintf(leaveMessage, client.GetName()),
		},
	}

	c.publishChatMessage(message.encode())
}

func (c *WsChat) GetID() string {
	return c.id.String()
}

func (c *WsChat) GetName() string {
	return c.name
}

func (c *WsChat) IsPrivate() bool {
	return c.private
}
