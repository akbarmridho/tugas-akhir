package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"tugas-akhir/backend/infrastructure/config"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

type QueryExecutor interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

const PostgresTransactionContextKey = "postgres_tx"

func (p *Postgres) GetExecutor(ctx context.Context) QueryExecutor {
	tx, ok := ctx.Value(PostgresTransactionContextKey).(pgx.Tx)

	if !ok {
		return p.Pool
	}

	return tx
}

// NewPostgres
// Example database url
// postgresql://username:password@leader.example.com:5432,follower1.example.com:5432,follower2.example.com:5432/dbname?target_session_attrs=primary
func NewPostgres(config *config.Config) (*Postgres, error) {
	c, err := pgxpool.ParseConfig(config.DatabaseUrl)
	c.MinConns = int32(config.DbMinPoolSize)

	pool, err := pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		Pool: pool,
	}, nil
}

var Module = fx.Options(fx.Provide(NewPostgres))
