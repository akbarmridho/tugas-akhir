package test_containers

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/yugabytedb"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
)

type RelationalDBVariant string

const (
	RelationalDBVariant__Postgres   RelationalDBVariant = "postgres"
	RelationalDBVariant__Citus      RelationalDBVariant = "citus"
	RelationalDBVariant__YugabyteDB RelationalDBVariant = "yugabytedb"
)

const TestDBName = "tugas-akhir"
const TestDBUser = "tugas-akhir"
const TestDBPassword = "tugas-akhir"
const TestYugabyteKeyspace = "tugas-akhir"

type RelationalDB struct {
	testcontainers.Container
	Variant RelationalDBVariant
}

func (r *RelationalDB) Cleanup(t testing.TB) {
	testcontainers.CleanupContainer(t, r.Container)
}

func NewRelationalDB(ctx context.Context, variant RelationalDBVariant) (*RelationalDB, error) {
	var container testcontainers.Container
	var initerr error

	if variant == RelationalDBVariant__Postgres {
		container, initerr = postgres.Run(ctx,
			"postgres:16.6.0",
			postgres.WithDatabase(TestDBName),
			postgres.WithUsername(TestDBUser),
			postgres.WithPassword(TestDBPassword),
			postgres.BasicWaitStrategies(),
		)
	} else if variant == RelationalDBVariant__Citus {
		req := testcontainers.ContainerRequest{
			Image:        "citusdata/citus:13.0.1-pg16",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       TestDBName,
				"POSTGRES_USER":     TestDBUser,
				"POSTGRES_PASSWORD": TestDBPassword,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections"),
		}
		container, initerr = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	} else if variant == RelationalDBVariant__YugabyteDB {
		container, initerr = yugabytedb.Run(
			ctx,
			"yugabytedb/yugabyte:2024.2.2.2-b2",
			yugabytedb.WithKeyspace(TestYugabyteKeyspace),
			yugabytedb.WithUser(TestDBUser),
			yugabytedb.WithDatabaseName(TestDBName),
			yugabytedb.WithDatabaseUser(TestDBUser),
			yugabytedb.WithDatabasePassword(TestDBPassword),
		)
	}

	if initerr != nil {
		return nil, initerr
	}

	return &RelationalDB{
		Container: container,
		Variant:   variant,
	}, nil
}
