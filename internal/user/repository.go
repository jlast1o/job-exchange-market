package user

import "context"

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
	SetBanned(ctx context.Context, id int64, banned bool, reason string) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*User, error)
}
