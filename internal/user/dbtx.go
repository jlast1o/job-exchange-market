package user

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX — минимальный набор методов для выполнения SQL-запросов.
// Его реализуют *pgxpool.Pool (обычное соединение) и pgx.Tx (транзакция).
// Благодаря этому репозиторий может работать с любым из них,
// не привязываясь к конкретному типу.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
