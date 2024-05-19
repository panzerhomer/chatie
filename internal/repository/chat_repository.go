package repository

import (
	"chatie/internal/models"
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
)

var ctx = context.Background()

type chatRepo struct {
	db *pgx.Conn
}

type ChatRepository interface {
	AddChat(ctx context.Context, chat *models.Chat) error
	AddUserToChat(ctx context.Context, user *models.User, chat *models.Chat) error
	GetChatByName(ctx context.Context, name string) (*models.Chat, error)
	GetChatByID(ctx context.Context, chatID string) (*models.Chat, error)
}

func NewChatRepository(db *pgx.Conn) ChatRepository {
	return &chatRepo{db: db}
}

func (r *chatRepo) AddChat(ctx context.Context, chat *models.Chat) error {
	query := "INSERT INTO chats(user_id, label, private, info) values($1, $2, $3, $4)"
	_, err := r.db.Exec(ctx, query, chat.GetId(), chat.GetName(), chat.GetPrivate())
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepo) AddUserToChat(ctx context.Context, user *models.User, chat *models.Chat) error {
	query := "INSERT INTO chat_members(chat_id, user_id, label, private, info) values($1, $2, $3, $4)"
	_, err := r.db.Exec(ctx, query, chat.GetId(), chat.Info, chat.GetPrivate())
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepo) GetChatByName(ctx context.Context, name string) (*models.Chat, error) {
	query := r.db.QueryRow(ctx, "SELECT chat_id, label, is_private FROM chats WHERE label = $1 LIMIT 1", name)

	var chat models.Chat

	if err := query.Scan(&chat.ID, &chat.Name, &chat.Private); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &chat, nil
}

func (r *chatRepo) GetChatByID(ctx context.Context, chatID string) (*models.Chat, error) {
	query := r.db.QueryRow(ctx, "SELECT chat_id, label, is_private FROM chats WHERE chat_id = $1 LIMIT 1", chatID)

	var chat models.Chat

	if err := query.Scan(&chat.ID, &chat.Name, &chat.Private); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &chat, nil
}
