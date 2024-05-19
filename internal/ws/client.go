package ws

import (
	"chatie/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 1 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 0,
}

// Client represents the websocket client at the server
type Client struct {
	conn     *websocket.Conn
	wsServer *WsServer
	send     chan []byte
	id       uuid.UUID
	userID   string
	name     string
	wsChats  map[*WsChat]bool
}

func newClient(conn *websocket.Conn, wsServer *WsServer, name string) *Client {
	return &Client{
		id:       uuid.New(),
		name:     name,
		conn:     conn,
		wsServer: wsServer,
		send:     make(chan []byte, 256),
		wsChats:  make(map[*WsChat]bool),
	}
}

func (client *Client) readPump() {
	defer func() {
		log.Println(client.wsServer.clients)
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		client.handleNewMessage(jsonMessage)
	}

}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The WsServer closed the channel.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
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
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	client.wsServer.unregister <- client
	for chat := range client.wsChats {
		chat.unregister <- client
	}
	close(client.send)
	client.conn.Close()
}

func ServeWS(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {
	name, ok := r.URL.Query()["name"]

	if !ok || len(name[0]) < 1 {
		log.Println("Url Param 'name' is missing")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, wsServer, name[0])

	go client.writePump()
	go client.readPump()

	wsServer.register <- client
}

func (client *Client) handleNewMessage(jsonMessage []byte) {
	var message WebsocketMessage
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
		return
	}

	message.Sender = clientToUser(client)

	log.Println("new event ", message.Action, message.Sender, message.Target, message.Message)

	switch message.Action {
	case SendMessageAction:
		client.handleSendMessage(message)
	case JoinChatAction:
		client.handleJoinChatMessage(message)
	case LeaveChatAction:
		client.handleLeaveChatMessage(message)
	case JoinChatPrivateAction:
		client.handleJoinChatPrivateMessage(message)
	case GetChatUsersAction:
		chatID := message.Target
		if chat := client.wsServer.findChatByName(chatID); chat != nil {
			if client.isInChat(chat) {
				data := SystemMessage{
					Action: "get-chat-users",
					Data:   chat.clients,
				}

				client.send <- data.encode()
			}
		}
	default:
		log.Println("unknown action", message.Action)
	}
}

func (client *Client) handleSendMessage(message WebsocketMessage) {
	chatID := message.Target
	if chat := client.wsServer.findChatByName(chatID); chat != nil {
		if client.isInChat(chat) {
			chat.broadcast <- &message
		} else {
			message := SystemMessage{
				Action: SendMessageAction,
				Data:   "illegal action",
			}
			client.send <- message.encode()
		}
	}
}

func (client *Client) handleJoinChatMessage(message WebsocketMessage) {
	chatName := message.Message.Text

	client.joinChat(chatName, nil)
}

func (client *Client) handleLeaveChatMessage(message WebsocketMessage) {
	chat := client.wsServer.findChatByName(message.Target)
	if chat == nil {
		return
	}

	delete(client.wsChats, chat)

	chat.unregister <- client
}

func (client *Client) handleJoinChatPrivateMessage(message WebsocketMessage) {
	target := client.wsServer.findClientByID(message.Target)

	if target == nil {
		return
	}

	chatName := message.Target + client.GetID()

	// client.joinChat(chatName, target)
	// target.joinRoom(chatName, client)
	joinedChat := client.joinChat(chatName, target)

	if joinedChat != nil {
		client.inviteTargetUser(target.GetName(), joinedChat)
	}
}

func (client *Client) joinChat(chatName string, sender *Client) *WsChat {
	chat := client.wsServer.findChatByName(chatName)
	if chat == nil {
		chat = client.wsServer.createChat(chatName, sender != nil)
	}

	if sender == nil && chat.IsPrivate() {
		return nil
	}

	if !client.isInChat(chat) {
		client.wsChats[chat] = true
		chat.register <- client

		client.notifyChatJoined(chat, sender)
	}

	return chat
}

func (client *Client) inviteTargetUser(targetID string, chat *WsChat) {
	inviteMessage := &WebsocketMessage{
		Action: JoinChatPrivateAction,
		Message: models.Message{
			Text: targetID,
		},
		Target: chat.GetID(),
		Sender: clientToUser(client),
	}

	if err := client.wsServer.redis.Publish(ctx, PubSubGeneralChannel, inviteMessage.encode()).Err(); err != nil {
		log.Println(err)
	}
}

func (client *Client) isInChat(chat *WsChat) bool {
	if _, ok := client.wsChats[chat]; ok {
		return true
	}

	return false
}

func (client *Client) notifyChatJoined(chat *WsChat, sender *Client) {
	message := WebsocketMessage{
		Action: ChatJoinedAction,
		Target: chat.GetID(),
		Sender: clientToUser(client),
	}

	client.send <- message.encode()
}

func (client *Client) GetName() string {
	return client.name
}

func (client *Client) GetID() string {
	return client.id.String()
}
