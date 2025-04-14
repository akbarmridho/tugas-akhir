package seeder

import (
	"testing"
	test_containers "tugas-akhir/backend/test-containers"
)

func TestSchemaManager_Postgres(t *testing.T) {
	t.Run("Postgres schema up", func(t *testing.T) {
		GetConnAndSchema(t, test_containers.RelationalDBVariant__Postgres)
	})
}

func TestSchemaManager_Citus(t *testing.T) {
	t.Run("Citus schema up", func(t *testing.T) {
		GetConnAndSchema(t, test_containers.RelationalDBVariant__Postgres)
	})
}

func TestSchemaManager_YugabyteDB(t *testing.T) {
	t.Run("YugabyteDB schema up", func(t *testing.T) {
		GetConnAndSchema(t, test_containers.RelationalDBVariant__Postgres)
	})
}
