package apperror

import (
	"errors"
)

var (
	ErrUserExists             = errors.New("user already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrUsersNotFound          = errors.New("users not found")
	ErrUserInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotUpdated         = errors.New("user not updated")
)

var (
	ErrInternal      = errors.New("internal error")
	ErrNotAuthorized = errors.New("not authorized")
)

var (
	ErrChatNotFound       = errors.New("chat not found")
	ErrChatsNotFound      = errors.New("chats not found")
	ErrChatNotUpdated     = errors.New("user not updated")
	ErrChatMemberNotFound = errors.New("chat member not found")
)
