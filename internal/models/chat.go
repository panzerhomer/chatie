package models

import "github.com/google/uuid"

// type WsChat struct {
// 	BaseModel
// 	Name     string    `json:"name"`
// 	Private  bool      `json:"private"`
// 	Info     string    `json:"info"`
// 	Members  []User    `json:"members"`
// 	Messages []Message `json:"members"`
// }

type Chat struct {
	BaseModel
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	Info     string `json:"info"`
	Link     string
	OwnerID  int
	Members  []User    `json:"members"`
	Messages []Message `json:"messages"`
}

func (chat *Chat) GenerateLink() {
	chat.Link = "join/" + uuid.New().String()
}

func (chat *Chat) GetId() int {
	return chat.ID
}

func (chat *Chat) GetName() string {
	return chat.Name
}

func (chat *Chat) GetPrivate() bool {
	return chat.Private
}
