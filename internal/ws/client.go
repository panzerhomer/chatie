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
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
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
	server *WsServer
	conn   *websocket.Conn
	send   chan []byte
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	rooms  map[*Room]bool
	sync.RWMutex
}

func newClient(conn *websocket.Conn, wsServer *WsServer, name string) *Client {
	return &Client{
		ID:     uuid.New(),
		Name:   name,
		conn:   conn,
		server: wsServer,
		send:   make(chan []byte, maxMessageSize),
		rooms:  make(map[*Room]bool),
	}

}

func (c *Client) disconnect() {
	c.server.unregister <- c
	for room := range c.rooms {
		room.unregister <- c
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
		log.Println("[readPump clicked by] ", c.Name)

		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.handleNewMessage(payload)
		// c.server.broadcast <- payload
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
			log.Println("[writePump clicked by] ", c.Name)

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
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

	if event.Payload == nil {
		return
	}

	newUser := clientToUser(client)
	event.Payload.From = newUser

	switch event.Type {
	case SendMessageEvent:
		client.handleSendMessage(event)
	case JoinRoomEvent:
		client.handleJoinRoomMessage(event)
	case LeaveRoomEvent:
		client.handleLeaveRoomMessage(event)
	case JoinRoomPrivateEvent:
		client.handleJoinRoomPrivateMessage(event)
	}
}

func (client *Client) handleSendMessage(event Event) {
	roomID := event.Payload.Room.Name
	if room := client.server.findRoomByName(roomID); room != nil {
		if client.isInRoom(room) {
			room.broadcast <- &event
		}
	}
}

func (client *Client) handleJoinRoomMessage(event Event) {
	roomName := event.Payload.Room.GetName()

	client.joinRoom(roomName, nil)
}

func (client *Client) handleLeaveRoomMessage(event Event) {
	room := client.server.findRoomByName(event.Payload.Room.Name)
	if room == nil {
		return
	}

	delete(client.rooms, room)

	room.unregister <- client
}

func (client *Client) handleJoinRoomPrivateMessage(message Event) {
	target := client.server.findClientByName(message.Payload.To.Name)

	if target == nil {
		return
	}

	roomName := target.Name + client.Name
	log.Println("[handleJoinRoomPrivateMessage]", roomName)

	client.joinRoom(roomName, target)
	target.joinRoom(roomName, client)

}

func (client *Client) joinRoom(roomName string, sender *Client) {
	room := client.server.findRoomByName(roomName)
	if room == nil {
		room = client.server.createRoom(roomName, sender != nil)
	}

	if sender == nil && room.Private {
		return
	}

	if !client.isInRoom(room) {
		client.rooms[room] = true
		room.register <- client
		client.notifyRoomJoined(room, sender)
	}
}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room]; ok {
		return true
	}
	return false
}

func (client *Client) notifyRoomJoined(room *Room, sender *Client) {
	message := Event{
		Type: RoomJoinedEvent,
		Payload: &WsMessage{
			Message: client.Name + " joined",
		},
	}

	client.send <- message.encode()
}
