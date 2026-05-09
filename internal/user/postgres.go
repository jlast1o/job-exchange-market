package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRepository struct {
	db DBTX
}

func NewPostgresRepository(db DBTX) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) WithTx(tx DBTX) *PostgresRepository {
	return &PostgresRepository{db: tx}
}

func (r *PostgresRepository) Create(ctx context.Context, u *User) error {
	query := `INSERT INTO users (email, password_hash, full_name, role) 
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, u.Email, u.PasswordHash, u.FullName, u.Role).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExist
		}

		return fmt.Errorf("Ошибка при создании пользователя: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id, email, password_hash, full_name, role, is_banned, banned_at, banned_reason, created_at, updated_at
	FROM users WHERE id = $1`
	return r.queryUserRow(ctx, query, id)
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password_hash, full_name, role, is_banned, banned_at, banned_reason, created_at, updated_at
	FROM users WHERE email = $1`
	return r.queryUserRow(ctx, query, email)
}

func (r *PostgresRepository) Update(ctx context.Context, u *User) error {
	query := `
	UPDATE users SET email = $2, password_hash = $3, full_name = $4, 
	role = $5, is_banned = $6, banned_at = $7, banned_reason = $8, updated_at = now()
	WHERE id = $1
	RETURNING updated_at `

	err := r.db.QueryRow(ctx, query, u.ID, u.Email, u.PasswordHash, u.FullName, u.Role, u.IsBanned, u.BannedAt, u.BannedReason).Scan(&u.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExist
		}

		return fmt.Errorf("Ошибка обновления юзера: %w", err)
	}

	return nil
}

func (r *PostgresRepository) SetBanned(ctx context.Context, id int64, banned bool, reason string) error {
	var bannedAt *time.Time

	if banned {
		now := time.Now()
		bannedAt = &now
	}

	query := `UPDATE users SET is_banned = $2, banned_at = $3, banned_reason = $4, updated_at = now() WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id, banned, bannedAt, reason)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}

		return fmt.Errorf("Ошибка бана/разбана пользователя: %w", err)
	}

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)

	return fmt.Errorf("Удаление пользователя: %w", err)
}

func (r *PostgresRepository) List(ctx context.Context, offset, limit int) ([]*User, error) {
	query := `SELECT id, email, password_hash, full_name, role, is_banned, banned_at, banned_reason, created_at, updated_at
	FROM users ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("Проблема получения списка пользователей: %w", err)
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		u := &User{}
		err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role,
			&u.IsBanned, &u.BannedAt, &u.BannedReason, &u.CreatedAt, &u.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("Проблема сканирования юзера: %w", err)
		}

		users = append(users, u)
	}

	return users, nil
}

func (r *PostgresRepository) queryUserRow(ctx context.Context, query string, args ...any) (*User, error) {
	u := &User{}

	err := r.db.QueryRow(ctx, query, args...).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName,
		&u.Role, &u.IsBanned, &u.BannedAt, &u.BannedReason, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("Ошибка получения пользователя: %w", err)
	}

	return u, nil
}
