package repository

import (
	"chatie/internal/apperror"
	"chatie/internal/models"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ctx = context.Background()

type chatRepo struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) *chatRepo {
	return &chatRepo{db: db}
}

func (r *chatRepo) Create(ctx context.Context, chat *models.Chat, userID int) (*models.Chat, error) {
	tx, err := r.db.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	queryUsers := `
		INSERT INTO 
			chats(owner_id, name, info, is_private, link) 
		VALUES ($1, $2, $3, $4, $5) RETURNING chat_id`

	_, err = tx.Prepare(ctx, "InsertIntoChats", queryUsers)
	if err != nil {
		return nil, err
	}

	var chatID int
	err = tx.QueryRow(ctx, "InsertIntoChats",
		userID, chat.Name, chat.Info, chat.Private, chat.Link).Scan(&chatID)
	if err != nil {
		if isDuplicateError(err) {
			return nil, apperror.ErrUserExists
		}
		return nil, err
	}

	queryChatMembers := `INSERT INTO chat_members(chat_id, user_id, user_role) VALUES ($1, $2, $3)`

	_, err = tx.Prepare(ctx, "InsertIntoMemmbers", queryChatMembers)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, "InsertIntoMembers", chat.ID, userID, models.UserOwner)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	chat.ID = chatID

	return chat, nil
}

func (r *chatRepo) Update(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	return nil, nil
}

func (r *chatRepo) AddUserToChat(ctx context.Context, chatID int, userID int, userRole string) error {
	query := `
		INSERT INTO 
			chat_members(chat_id, user_id, user_role)
		VALUES
			($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, chatID, userID, userRole)
	if err != nil {
		return err
	}

	return nil
}

func (r *chatRepo) GetChatMemberByID(ctx context.Context, chatID int, userID int) (*models.ChatUser, error) {
	query := `
		SELECT 
			u.user_id, 
			u.username,
			u.firstname,
			u.lastname,
			u.patronymic,
			u.info,
			u.email,
			cm.user_role,
			cm.is_banned,
			cm.joined_at
		FROM
			users as u
		JOIN
			chat_members as cm
		ON
			u.user_id = cm.user_id
		WHERE cm.chat_id = $2 AND cm.user_id = $1`

	var chatUser models.ChatUser

	err := r.db.QueryRow(ctx, query, userID, chatID).Scan(
		&chatUser.ID,
		&chatUser.Username,
		&chatUser.Name,
		&chatUser.Lastname,
		&chatUser.Patronymic,
		&chatUser.Info,
		&chatUser.Email,
		&chatUser.Role,
		&chatUser.IsBanned,
		&chatUser.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrChatMemberNotFound
		}
		return nil, err
	}

	return &chatUser, nil

}

func (r *chatRepo) GetChatMembersByID(ctx context.Context, chatID int) ([]models.ChatUser, error) {
	query := `
		SELECT 
			u.user_id, 
			u.username,
			u.firstname,
			u.lastname,
			u.patronymic,
			u.info,
			u.email,
			cm.user_role,
			cm.is_banned,
			cm.joined_at
		FROM
			users AS u
		JOIN
			chat_members AS cm
		ON
			u.user_id = cm.user_id
		WHERE 
			cm.chat_id = $1`

	rows, err := r.db.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chatUsers []models.ChatUser

	for rows.Next() {
		var chatUser models.ChatUser
		err := rows.Scan(
			&chatUser.ID,
			&chatUser.Username,
			&chatUser.Name,
			&chatUser.Lastname,
			&chatUser.Patronymic,
			&chatUser.Info,
			&chatUser.Email,
			&chatUser.Role,
			&chatUser.IsBanned,
			&chatUser.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		chatUsers = append(chatUsers, chatUser)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return chatUsers, nil
}

func (r *chatRepo) GetByID(ctx context.Context, chatID int) (*models.Chat, error) {
	query := `
		SELECT
			chat_id,
			owner_id, 
			name, 
			info, 
			is_private, 
			link,
			created_at
		FROM
			chats
		WHERE
			chat_id = $1`

	var chat models.Chat

	err := r.db.QueryRow(ctx, query, chatID).Scan(
		&chat.ID,
		&chat.OwnerID,
		&chat.Name,
		&chat.Info,
		&chat.Private,
		&chat.Link,
		&chat.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrChatMemberNotFound
		}
		return nil, err
	}

	return &chat, nil

}

func (r *chatRepo) GetAllChatsByUserID(ctx context.Context, userID int) ([]models.Chat, error) {
	query := `
		SELECT
			chat_id,
			name, 
			info, 
			is_private, 
			link,
			created_at
		FROM
			chats
		WHERE
			owner_id = $1`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		err = rows.Scan(
			&chat.ID,
			&chat.Name,
			&chat.Info,
			&chat.Link,
			&chat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}
