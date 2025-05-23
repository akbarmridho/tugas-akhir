package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/fx"
	"time"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/pkg/logger"
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
	c.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	l := logger.GetDebug()

	if config.EnableDBTracing {

		tlogger := NewZapTracer(l)

		tracer := tracelog.TraceLog{
			Logger:   tlogger,
			LogLevel: tracelog.LogLevelTrace,
		}

		c.ConnConfig.Tracer = &tracer

		l.Info("Database tracing enabled")
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return nil, err
	}

	if config.LogPoolStat {
		l.Info("Log pool stat")

		go func(p *pgxpool.Pool) {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					currentStats := p.Stat()
					fmt.Printf("\n--- Pool Stats")
					fmt.Printf("  AcquireCount: %d\n", currentStats.AcquireCount())
					fmt.Printf("  AcquireDuration: %s\n", currentStats.AcquireDuration())
					fmt.Printf("  AcquiredConns: %d\n", currentStats.AcquiredConns())
					fmt.Printf("  CanceledAcquireCount: %d\n", currentStats.CanceledAcquireCount())
					fmt.Printf("  ConstructingConns: %d\n", currentStats.ConstructingConns())
					fmt.Printf("  EmptyAcquireCount: %d\n", currentStats.EmptyAcquireCount())
					fmt.Printf("  IdleConns: %d\n", currentStats.IdleConns())
					fmt.Printf("  MaxConns: %d\n", currentStats.MaxConns())
					fmt.Printf("  TotalConns: %d\n", currentStats.TotalConns())
					fmt.Printf("  NewConnsCount: %d\n", currentStats.NewConnsCount())
					fmt.Printf("  MaxLifetimeDestroyCount: %d\n", currentStats.MaxLifetimeDestroyCount())
					fmt.Printf("  MaxIdleDestroyCount: %d\n", currentStats.MaxIdleDestroyCount())
					fmt.Println("-----------------------------------")
				}
			}
		}(pool)
	}

	return &Postgres{
		Pool: pool,
	}, nil
}

var Module = fx.Options(fx.Provide(NewPostgres))
