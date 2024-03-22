package ws

import (
	"chatie/internal/models"
	"encoding/json"
	"log"
)

const (
	ChatJoinedEvent = "chat-joined"
	UserLeftEvent   = "user-left"
	UserJoinedEvent = "user-join"
)

const SendMessageEvent = "send-message"
const JoinChatEvent = "join-chat"
const LeaveChatEvent = "leave-chat"
const JoinChatPrivateEvent = "join-chat-private"

type Event struct {
	Type    string     `json:"type"`
	Payload *WsMessage `json:"payload"`
}

// message from client
type WsMessage struct {
	Chat    *Chat  `json:"chat"`
	Message string `json:"message"`
	// From    *Client `json:"from"`
	// To      *Client `json:"to"`
	From *models.User `json:"from"`
	To   *models.User `json:"to"`
}

func (e *Event) encode() []byte {
	encoding, err := json.Marshal(e)
	if err != nil {
		log.Println(err)
	}

	return encoding
}
