package models

import "errors"

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrInternal     = errors.New("internal error")
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)
