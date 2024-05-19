package models

type WsChat struct {
	BaseModel
	Name     string    `json:"name"`
	Private  bool      `json:"private"`
	Info     string    `json:"info"`
	Members  []User    `json:"members"`
	Messages []Message `json:"members"`
}

type Chat struct {
	BaseModel
	Name     string    `json:"name"`
	Private  bool      `json:"private"`
	Info     string    `json:"info"`
	Members  []User    `json:"members"`
	Messages []Message `json:"messages"`
}

func (chat *Chat) GetId() string {
	return chat.ID
}

func (chat *Chat) GetName() string {
	return chat.Name
}

func (chat *Chat) GetPrivate() bool {
	return chat.Private
}
