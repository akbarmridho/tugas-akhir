package seeder

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/postgres"
	test_containers "tugas-akhir/backend/test-containers"
)

func TestSchemaManager_Postgres(t *testing.T) {
	t.Run("Postgres schema up", func(t *testing.T) {
		ctx := t.Context()
		container, err := test_containers.NewRelationalDB(ctx, test_containers.RelationalDBVariant__Postgres)
		require.NoError(t, err)
		container.Cleanup(t)

		host, err := container.Host(ctx)
		require.NoError(t, err)

		port, err := container.MappedPort(ctx, "5432/tcp")
		require.NoError(t, err)

		conf := config.Config{
			DatabaseUrl: fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", test_containers.TestDBUser, test_containers.TestDBPassword, host, port.Port(), test_containers.TestDBName),
		}

		conn, err := postgres.NewPostgres(&conf)
		require.NoError(t, err)

		schemaManager := NewSchemaManager(conn)

		require.NoError(t, schemaManager.SchemaUp(ctx))
	})
}

func TestSchemaManager_Citus(t *testing.T) {
	t.Run("Citus schema up", func(t *testing.T) {
		ctx := t.Context()
		container, err := test_containers.NewRelationalDB(ctx, test_containers.RelationalDBVariant__Citus)
		require.NoError(t, err)
		container.Cleanup(t)

		host, err := container.Host(ctx)
		require.NoError(t, err)

		port, err := container.MappedPort(ctx, "5432/tcp")
		require.NoError(t, err)

		conf := config.Config{
			DatabaseUrl: fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", test_containers.TestDBUser, test_containers.TestDBPassword, host, port.Port(), test_containers.TestDBName),
		}

		conn, err := postgres.NewPostgres(&conf)
		require.NoError(t, err)

		schemaManager := NewSchemaManager(conn)

		require.NoError(t, schemaManager.SchemaUp(ctx))
		require.NoError(t, schemaManager.CitusSetup(ctx))
	})
}

func TestSchemaManager_YugabyteDB(t *testing.T) {
	t.Run("YugabyteDB schema up", func(t *testing.T) {
		ctx := t.Context()
		container, err := test_containers.NewRelationalDB(ctx, test_containers.RelationalDBVariant__YugabyteDB)
		require.NoError(t, err)
		container.Cleanup(t)

		host, err := container.Host(ctx)
		require.NoError(t, err)

		port, err := container.MappedPort(ctx, "5433/tcp")
		require.NoError(t, err)

		conf := config.Config{
			DatabaseUrl: fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", test_containers.TestDBUser, test_containers.TestDBPassword, host, port.Port(), test_containers.TestDBName),
		}

		conn, err := postgres.NewPostgres(&conf)
		require.NoError(t, err)

		schemaManager := NewSchemaManager(conn)

		require.NoError(t, schemaManager.SchemaUp(ctx))
	})
}
