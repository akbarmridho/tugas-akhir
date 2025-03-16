package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Risingwave struct {
	Pool *pgxpool.Pool
}

// NewRisingwave
// Example database url
func NewRisingwave(config Config) (*Risingwave, error) {
	c, err := pgxpool.ParseConfig(config.DatabaseUrl)

	pool, err := pgxpool.NewWithConfig(context.Background(), c)
	if err != nil {
		return nil, err
	}

	return &Risingwave{
		Pool: pool,
	}, nil
}
