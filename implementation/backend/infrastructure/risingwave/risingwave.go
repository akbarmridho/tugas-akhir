package risingwave

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"tugas-akhir/backend/infrastructure/config"
)

type Risingwave struct {
	Pool *pgxpool.Pool
}

// NewRisingwave
// Example database url
func NewRisingwave(config *config.Config) (*Risingwave, error) {
	c, err := pgxpool.ParseConfig(config.RisingwaveUrl)

	pool, err := pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return nil, err
	}

	return &Risingwave{
		Pool: pool,
	}, nil
}

var Module = fx.Options(fx.Provide(NewRisingwave))
