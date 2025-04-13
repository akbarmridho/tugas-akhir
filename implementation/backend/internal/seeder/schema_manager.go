package seeder

import (
	"context"
	_ "embed"
	"strings"
	"tugas-akhir/backend/infrastructure/postgres"
)

//go:embed schemas/schema_up.sql
var SchemaUp string

//go:embed schemas/schema_down.sql
var SchemaDown string

//go:embed schemas/citus_setup.sql
var CitusSetup string

type SchemaManager struct {
	db *postgres.Postgres
}

func NewSchemaManager(db *postgres.Postgres) *SchemaManager {
	return &SchemaManager{
		db: db,
	}
}

func (m *SchemaManager) SchemaUp(ctx context.Context) error {
	_, err := m.db.Pool.Exec(ctx, SchemaUp)
	return err
}

func (m *SchemaManager) SchemaDown(ctx context.Context) error {
	_, err := m.db.Pool.Exec(ctx, SchemaDown)
	return err
}

func (m *SchemaManager) CitusSetup(ctx context.Context) error {
	for _, cmd := range strings.Split(CitusSetup, "-- marker split") {
		_, err := m.db.Pool.Exec(ctx, cmd)
		return err
	}

	return nil
}
