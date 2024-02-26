package ws

import (
	"encoding/json"
	"log"
)

const SendMessageEvent = "send-message"
const JoinRoomEvent = "join-room"
const LeaveRoomEvent = "leave-room"
const UserJoinedEvent = "user-join"
const UserLeftEvent = "user-left"
const JoinRoomPrivateEvent = "join-room-private"
const RoomJoinedEvent = "room-joined"

type Event struct {
	Type    string           `json:"type"`
	Payload *ReceivedMessage `json:"payload"`
}

// message from a client
type ReceivedMessage struct {
	Room    *Room   `json:"room"`
	Message string  `json:"message"`
	From    *Client `json:"from"`
	To      *Client `json:"to"`
}

// message from system
type SentMessage struct {
	Type string `json:"type"`
	// Data any    `json:"data"`
	From *Client `json:"from"`
	Chat *Room   `json:"room"`
}

func (e *Event) encode() []byte {
	encoding, err := json.Marshal(e)
	if err != nil {
		log.Println(err)
	}

	return encoding
}
