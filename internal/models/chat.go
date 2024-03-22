package models

type Chat struct {
	ID       string `json:"chat_id"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	Members  []User
	Messages []Message
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
