package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

type Client struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	server *WsServer
	conn   *websocket.Conn
	send   chan []byte
	Chats  map[*Chat]bool
	sync.RWMutex
}

func newClient(conn *websocket.Conn, wsServer *WsServer, name string) *Client {
	return &Client{
		ID:     uuid.New(),
		Name:   name,
		conn:   conn,
		server: wsServer,
		send:   make(chan []byte, maxMessageSize),
		Chats:  make(map[*Chat]bool),
	}

}

func (c *Client) disconnect() {
	c.server.unregister <- c
	for Chat := range c.Chats {
		Chat.unregister <- c
	}
	close(c.send)
	c.conn.Close()
}

func (client *Client) GetName() string {
	return client.Name
}

func (c *Client) readPump() {
	defer func() {
		c.disconnect()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.handleNewMessage(payload)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.disconnect()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeHTTP(server *WsServer, w http.ResponseWriter, r *http.Request) {
	name, ok := r.URL.Query()["name"]

	if !ok || len(name[0]) < 1 {
		log.Println("name is missed")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, server, name[0])
	server.register <- client

	go client.writePump()
	go client.readPump()
}

func (client *Client) handleNewMessage(jsonMessage []byte) {
	var event Event
	if err := json.Unmarshal(jsonMessage, &event); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
		return
	}

	log.Println("[handleNewMessage]", event.Type, event.Payload)

	if event.Payload == nil {
		return
	}

	newUser := clientToUser(client)
	event.Payload.From = newUser

	switch event.Type {
	case SendMessageEvent:
		client.handleSendMessage(event)
	case JoinChatEvent:
		client.handleJoinChatMessage(event)
	case LeaveChatEvent:
		client.handleLeaveChatMessage(event)
	case JoinChatPrivateEvent:
		client.handleJoinChatPrivateMessage(event)
	default:
		client.conn.WriteJSON("wrong event")
	}
}

func (client *Client) handleSendMessage(event Event) {
	сhat := event.Payload.Chat.Name
	if chat := client.server.findChatByName(сhat); chat != nil {
		if client.isInChat(chat) {
			chat.broadcast <- &event
		}
	}
}

func (client *Client) handleJoinChatMessage(event Event) {
	chatName := event.Payload.Chat.GetName()

	client.joinChat(chatName, nil)
}

func (client *Client) handleLeaveChatMessage(event Event) {
	сhat := client.server.findChatByName(event.Payload.Chat.Name)
	if сhat == nil {
		return
	}

	delete(client.Chats, сhat)

	сhat.unregister <- client
}

func (client *Client) handleJoinChatPrivateMessage(message Event) {
	target := client.server.findClientByName(message.Payload.To.Name)
	if target == nil {
		return
	}

	chatName := target.Name + client.Name

	client.joinChat(chatName, target)
	target.joinChat(chatName, client)
}

func (client *Client) joinChat(ChatName string, sender *Client) {
	chat := client.server.findChatByName(ChatName)
	if chat == nil {
		chat = client.server.createChat(ChatName, sender != nil)
	}

	if sender == nil && chat.Private {
		return
	}

	if !client.isInChat(chat) {
		client.Chats[chat] = true
		chat.register <- client
		client.notifyChatJoined(chat, sender)
	}
}

func (client *Client) isInChat(chat *Chat) bool {
	if _, ok := client.Chats[chat]; ok {
		return true
	}
	return false
}

func (client *Client) notifyChatJoined(chat *Chat, sender *Client) {
	message := Event{
		Type: ChatJoinedEvent,
		Payload: &WsMessage{
			Message: client.Name + " joined",
		},
	}

	client.send <- message.encode()
}
