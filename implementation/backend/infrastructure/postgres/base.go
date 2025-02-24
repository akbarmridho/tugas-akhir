package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Base struct {
	Pool *pgxpool.Pool
}

type QueryExecutor interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

const TransactionContextKey = "postgres_tx"

func (p *Base) GetExecutor(ctx context.Context) QueryExecutor {
	tx, ok := ctx.Value(TransactionContextKey).(pgx.Tx)

	if !ok {
		return p.Pool
	}

	return tx
}

func NewBasePostgres(config Config) (*Base, error) {
	c, err := pgxpool.ParseConfig(fmt.Sprintf(
		"user=%s password=%s host=%s port=%s database=%s",
		config.DatabaseUsername,
		config.DatabasePassword,
		config.DatabaseHost,
		config.DatabasePort,
		config.DatabaseName,
	))

	// todo adjust the max conn parameter

	if config.Timezone != nil {
		c.ConnConfig.RuntimeParams["timezone"] = *config.Timezone
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return nil, err
	}

	return &Base{
		Pool: pool,
	}, nil
}
