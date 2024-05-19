package models

type Message struct {
	BaseModel
	Text       string      `json:"text"`
	UserID     string      `json:"userID"`
	ChatID     string      `json:"chatID"`
	ChannelID  string      `json:"channelID"`
	Attachment *Attachment `json:"attachment"`
}

type Attachment struct {
	BaseModel
	Url       string `json:"url"`
	FileType  string `json:"filetype"`
	Filename  string `json:"filename"`
	MessageID string `json:"-"`
}
