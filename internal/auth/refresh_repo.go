package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRefreshTokenNotFound = errors.New("Рефреш токен не найден")

type RefreshTokenRepository interface {
	Store(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	Validate(ctx context.Context, tokenHash string) error
	Revoke(ctx context.Context, tokenHash string)
	DeleteExpired(ctx context.Context) error
}

type PostgresRefreshTokenRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRefreshTokenRepo(pool *pgxpool.Pool) *PostgresRefreshTokenRepo {
	return &PostgresRefreshTokenRepo{
		pool: pool,
	}
}

func (r *PostgresRefreshTokenRepo) Store(ctx context.Context, userID int64, tokenHash string, expiredAt time.Time) error {
	const query = `INSERT INTO refresh_tokens (user_id, token_hash, expired_at) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, userID, tokenHash, expiredAt)

	return fmt.Errorf("Установка рефреш токена: %w", err)
}

func (r *PostgresRefreshTokenRepo) Validate(ctx context.Context, tokenHash string) error {
	var expiresAt time.Time
	const query = `SELECT expires_at FROM refresh_tokens WHERE token_hash = $1`
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&expiresAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrRefreshTokenNotFound
		}

		return fmt.Errorf("Валидация рефреш токена: %w", err)
	}

	if time.Now().After(expiresAt) {
		return ErrRefreshTokenNotFound
	}

	return nil
}

func (r *PostgresRefreshTokenRepo) Revoke(ctx context.Context, tokenHash string) error {
	const query = `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := r.pool.Exec(ctx, query, tokenHash)

	return fmt.Errorf("Удаление рефреш ключа: %w", err)
}

func (r *PostgresRefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	const query = `DELETE FROM refresh_tokens WHERE expires_at < now()`
	_, err := r.pool.Exec(ctx, query)

	return fmt.Errorf("Удаление просроченных ключей: %w", err)
}
