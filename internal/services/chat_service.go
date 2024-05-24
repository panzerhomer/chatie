package services

import (
	"chatie/internal/models"
	"context"
	"errors"
)

type ChatRepository interface {
	Create(ctx context.Context, chat *models.Chat, userID int) (*models.Chat, error)
	Update(ctx context.Context, chat *models.Chat) (*models.Chat, error)
	GetByID(ctx context.Context, chatID int) (*models.Chat, error)

	AddUserToChat(ctx context.Context, chatID int, userID int, userRole string) error
	GetChatMemberByID(ctx context.Context, chatID int, userID int) (*models.ChatUser, error)
	GetChatMembersByID(ctx context.Context, chatID int) ([]models.ChatUser, error)
	GetAllChatsByUserID(ctx context.Context, userID int) ([]models.Chat, error)
}

type chatService struct {
	repo ChatRepository
}

func (c *chatService) AddChat(ctx context.Context, chat models.Chat, userID int) (*models.Chat, error) {
	createdChat, err := c.repo.Create(ctx, &chat, userID)
	if err != nil {
		return nil, err
	}

	return createdChat, nil
}

func (c *chatService) JoinChat(ctx context.Context, chatID int, userID int) error {
	err := c.repo.AddUserToChat(ctx, chatID, userID, models.UserDefault)
	if err != nil {
		return err
	}

	return nil
}

func (c *chatService) IsOwner(ctx context.Context, chatID int, userID int) bool {
	chat, err := c.repo.GetByID(ctx, chatID)
	if err != nil {
		return false
	}

	return chat.OwnerID == userID
}

func (c *chatService) GetAllUserChats(ctx context.Context, userID int) ([]models.Chat, error) {
	chats, err := c.repo.GetAllChatsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return chats, nil
}

func (c *chatService) checkRole(ctx context.Context, chatID int, userID int) (string, error) {
	user, err := c.repo.GetChatMemberByID(ctx, chatID, userID)
	if err != nil {
		return "", err
	}

	if user.Role == "" {
		return "", errors.New("no roles")
	}

	return user.Role, nil
}

func (c *chatService) IsAdmin(ctx context.Context, chatID, userID int) bool {
	role, err := c.checkRole(ctx, chatID, userID)
	if err != nil {
		return false
	}

	return role == models.UserAdmin
}
