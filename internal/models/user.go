package models

type User struct {
	BaseModel
	Role     string    `json:"role"`
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
	IsOnline bool      `json:"isOnline"`
	Messages []Message `json:"-"`
}

func (u *User) GetID() string {
	return u.ID
}

func (u *User) GetName() string {
	return u.Username
}
