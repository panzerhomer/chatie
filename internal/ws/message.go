package ws

import (
	"chatie/internal/models"
	"encoding/json"
	"log"
)

// Subscribed Messages
const SendMessageAction = "send-message"
const JoinChatAction = "join-chat"
const LeaveChatAction = "leave-chat"
const UserJoinedAction = "user-joined"
const UserLeftAction = "user-left"
const JoinChatPrivateAction = "join-chat-private"

const ChatJoinedAction = "chat-joined"
const GetChatUsersAction = "get-chat-users"

type WebsocketMessage struct {
	Action  string          `json:"action"`
	Message *models.Message `json:"message"`
	Target  string          `json:"target"` // chat, user, channel ids
	Sender  *models.User    `json:"sender"`
}

type SystemMessage struct {
	Action string `json:"action"`
	Data   any    `json:"data"`
}

func (message *SystemMessage) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}

func (message *WebsocketMessage) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}
