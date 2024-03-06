package ws

import (
	"chatie/internal/models"
	"encoding/json"
	"log"
)

const (
	RoomJoinedEvent = "room-joined"
	UserLeftEvent   = "user-left"
	UserJoinedEvent = "user-join"
)

const SendMessageEvent = "send-message"
const JoinRoomEvent = "join-room"
const LeaveRoomEvent = "leave-room"
const JoinRoomPrivateEvent = "join-room-private"

type Event struct {
	Type    string     `json:"type"`
	Payload *WsMessage `json:"payload"`
}

// message from a client
type WsMessage struct {
	Room    *Room  `json:"room"`
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
