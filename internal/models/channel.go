package models

type Channel struct {
	BaseModel
	Name     string
	Private  bool
	Members  []User
	Messages []Message
}
