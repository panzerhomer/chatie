package models

import "time"

// describes message in a room
type Message struct {
	BaseModel
	Text       *string
	UserID     string
	RoomID     string
	Attachment *Attachment
}

type Attachment struct {
	ID        string    `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Url       string    `json:"url"`
	FileType  string    `json:"filetype"`
	Filename  string    `json:"filename"`
	MessageID string    `json:"-"`
}
