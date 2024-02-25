package models

import (
	"time"
)

type BaseModel struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Success struct {
	Success bool `json:"success"`
}
