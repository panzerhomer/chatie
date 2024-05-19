package apperror

import (
	"errors"
)

var (
	ErrUserExists             = errors.New("user already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrUsersNotFound          = errors.New("users not found")
	ErrUserInvalidCredentials = errors.New("invalid credentials")
)

var (
	ErrInternal      = errors.New("internal error")
	ErrNotAuthorized = errors.New("not authorized")
)
