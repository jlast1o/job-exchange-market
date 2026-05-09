package user

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("Пользователь не найден")
	ErrAlreadyExist = errors.New("Пользователь уже существует")
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	FullName     string
	Role         string
	IsBanned     bool
	BannedAt     *time.Time
	BannedReason *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
