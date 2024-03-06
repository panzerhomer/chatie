package models

type Room struct {
	ID      string
	Name    string
	Private bool
}

func (room *Room) GetId() string {
	return room.ID
}

func (room *Room) GetName() string {
	return room.Name
}

func (room *Room) GetPrivate() bool {
	return room.Private
}
