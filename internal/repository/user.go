package repository

import (
	"context"
	"time"
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

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	SetBanned(ctx context.Context, id int64, banned bool, reason string) error
	SetUnbanned(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*User, error)
}
