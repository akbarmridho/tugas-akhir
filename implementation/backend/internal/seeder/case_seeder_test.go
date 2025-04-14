package seeder

import (
	"testing"
	test_containers "tugas-akhir/backend/test-containers"
)

func TestSeeder_Postgres(t *testing.T) {
	t.Run("Postgres schema up", func(t *testing.T) {
		conn := GetConnAndSchema(t, test_containers.RelationalDBVariant__Postgres)
		SeedSchema(t, t.Context(), conn)
	})
}

func TestSeeder_Citus(t *testing.T) {
	t.Run("Citus schema up", func(t *testing.T) {
		conn := GetConnAndSchema(t, test_containers.RelationalDBVariant__Citus)
		SeedSchema(t, t.Context(), conn)
	})
}

func TestSeeder_YugabyteDB(t *testing.T) {
	t.Run("YugabyteDB schema up", func(t *testing.T) {
		conn := GetConnAndSchema(t, test_containers.RelationalDBVariant__YugabyteDB)
		SeedSchema(t, t.Context(), conn)
	})
}
