package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRefreshTokenNotFound = errors.New("Рефреш токен не найден")

type RefreshTokenRepository interface {
	Store(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	Validate(ctx context.Context, tokenHash string) error
	Revoke(ctx context.Context, tokenHash string) error
	DeleteExpired(ctx context.Context) error

	WithTx(db DBTX) RefreshTokenRepository
}

type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresRefreshTokenRepo struct {
	db DBTX
}

func NewPostgresRefreshTokenRepo(pool *pgxpool.Pool) *PostgresRefreshTokenRepo {
	return &PostgresRefreshTokenRepo{db: pool}
}

func (r *PostgresRefreshTokenRepo) WithTx(db DBTX) RefreshTokenRepository {
	return &PostgresRefreshTokenRepo{db: db}
}

func (r *PostgresRefreshTokenRepo) Store(ctx context.Context, userID int64, tokenHash string, expiredAt time.Time) error {
	const query = `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiredAt)

	if err != nil {
		return fmt.Errorf("Установка рефреш токена: %w", err)
	}
	return nil
}
func (r *PostgresRefreshTokenRepo) Validate(ctx context.Context, tokenHash string) error {
	var expiresAt time.Time
	const query = `SELECT expires_at FROM refresh_tokens WHERE token_hash = $1`
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&expiresAt)

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
	_, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("Удаление рефреш ключа: %w", err)
	}
	return nil
}

func (r *PostgresRefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	const query = `DELETE FROM refresh_tokens WHERE expires_at < now()`
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("Удаление просроченных ключей: %w", err)
	}
	return nil
}
