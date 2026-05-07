package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		pool: pool,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *User) error {
	query := `INSERT INTO users (email, password_hash, full_name, role) values ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, u.Email, u.PasswordHash, u.FullName, u.Role).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	return fmt.Errorf("Создание юзера: %w", err)
}
