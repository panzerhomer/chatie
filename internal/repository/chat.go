package repository

import (
	"chatie/internal/models"
	"context"
	"database/sql"
)

type chatRepo struct {
	db *sql.DB
}

type ChatRepository interface {
	AddChat(ctx context.Context, chat *models.Chat) error
	GetChatByName(ctx context.Context, name string) (*models.Chat, error)
}

func NewChatRepository(db *sql.DB) ChatRepository {
	return &chatRepo{db: db}
}

func (repo *chatRepo) AddChat(ctx context.Context, chat *models.Chat) error {
	query, err := repo.db.Prepare("INSERT INTO chat(id, name, private) values(?,?,?)")
	if err != nil {
		return err
	}

	_, err = query.Exec(chat.GetId(), chat.GetName(), chat.GetPrivate())
	if err != nil {
		return err
	}
	return nil
}

func (repo *chatRepo) GetChatByName(ctx context.Context, name string) (*models.Chat, error) {
	query := repo.db.QueryRow("SELECT id, name, private FROM chat where name = ? LIMIT 1", name)

	var chat models.Chat

	if err := query.Scan(&chat.ID, &chat.Name, &chat.Private); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &chat, nil

}
