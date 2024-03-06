package models

type Message struct {
	BaseModel
	Text       *string
	UserID     string
	RoomID     string
	Attachment *Attachment
}

type Attachment struct {
	BaseModel
	Url       string `json:"url"`
	FileType  string `json:"filetype"`
	Filename  string `json:"filename"`
	MessageID string `json:"-"`
}
