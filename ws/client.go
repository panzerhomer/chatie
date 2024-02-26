package ws

import (
	"encoding/json"
	"log"
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
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	ID         uuid.UUID `json:"id"`
	server     *WsServer
	connection *websocket.Conn
	send       chan []byte
	rooms      map[*Room]bool
	Name       string `json:"name"`
}

func newClient(conn *websocket.Conn, server *WsServer, name string) *Client {
	return &Client{
		ID:         uuid.New(),
		Name:       name,
		connection: conn,
		server:     server,
		send:       make(chan []byte, 256),
		rooms:      make(map[*Room]bool),
	}
}

func (client *Client) GetName() string {
	return client.Name
}

func (client *Client) isInRoom(room *Room) bool {
	if _, ok := client.rooms[room]; ok {
		return true
	}
	return false
}

func (client *Client) disconnect() {
	client.server.unregister <- client
	for room := range client.rooms {
		room.unregister <- client
	}
	close(client.send)
	client.connection.Close()
}

func (client *Client) readPump() {
	defer func() {
		client.disconnect()
	}()

	client.connection.SetReadLimit(maxMessageSize)
	client.connection.SetReadDeadline(time.Now().Add(pongWait))
	client.connection.SetPongHandler(func(string) error { client.connection.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, payload, err := client.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		// reading incoming events
		var incomingEvent Event
		if err := json.Unmarshal(payload, &incomingEvent); err != nil {
			log.Printf("error marshalling message: %v", err)
			break
		}

		log.Println("[client incomEvent]", incomingEvent.Type, incomingEvent.Payload)

		if err := client.server.handleNewMessage(client, &incomingEvent); err != nil {
			log.Println("error handeling message: ", err)
		}
		// client.server.broadcast <- payload
	}
}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.disconnect()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
