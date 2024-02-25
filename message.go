package main

import (
	"encoding/json"
	"log"
)

const SendMessageAction = "send-message"
const JoinRoomAction = "join-room"
const LeaveRoomAction = "leave-room"
const UserJoinedAction = "user-join"
const UserLeftAction = "user-left"
const JoinRoomPrivateAction = "join-room-private"
const RoomJoinedAction = "room-joined"

type Event struct {
	Type    string           `json:"type"`
	Payload *ReceivedMessage `json:"payload"`
}

type ReceivedMessage struct {
	Room    string `json:"room"`
	Message string `json:"message"`
}

type WebsocketMessage struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func (message *Event) encode() []byte {
	encoding, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return encoding
}
