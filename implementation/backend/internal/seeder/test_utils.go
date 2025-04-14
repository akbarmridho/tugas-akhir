package seeder

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"testing"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/postgres"
	test_containers "tugas-akhir/backend/test-containers"
)

func GetConnAndSchema(t *testing.T, variant test_containers.RelationalDBVariant) *postgres.Postgres {
	ctx := t.Context()
	container, err := test_containers.NewRelationalDB(ctx, variant)
	require.NoError(t, err)
	container.Cleanup(t)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	var port nat.Port

	if variant == test_containers.RelationalDBVariant__YugabyteDB {
		port, err = container.MappedPort(ctx, "5433/tcp")
		require.NoError(t, err)
	} else {
		port, err = container.MappedPort(ctx, "5432/tcp")
		require.NoError(t, err)
	}

	conf := config.Config{
		DatabaseUrl: fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", test_containers.TestDBUser, test_containers.TestDBPassword, host, port.Port(), test_containers.TestDBName),
	}

	conn, err := postgres.NewPostgres(&conf)
	require.NoError(t, err)

	schemaManager := NewSchemaManager(conn)

	require.NoError(t, schemaManager.SchemaUp(ctx))

	if variant == test_containers.RelationalDBVariant__Citus {
		require.NoError(t, schemaManager.CitusSetup(ctx))
	}

	return conn
}

func SeedSchema(t *testing.T, ctx context.Context, conn *postgres.Postgres) {
	seeder := NewCaseSeeder(conn)

	payload := SeederPayload{
		DayCount: 2,
		SeatedCategories: []CategoryPayload{
			{Name: "Grandstand Row A-E", Price: 750000, AreaCount: 1, SeatPerArea: 100},
			{Name: "Grandstand Row F-M", Price: 500000, AreaCount: 1, SeatPerArea: 250},
		},
		FreeStandingCategories: []CategoryPayload{
			{Name: "Front Stage Pit", Price: 600000, AreaCount: 1, SeatPerArea: 300},
			{Name: "General Lawn Area", Price: 300000, AreaCount: 2, SeatPerArea: 1000},
		},
	}

	require.NoError(t, seeder.Seed(ctx, payload))

	// Check for inserted rows from various tables

	// 1. Check events table - should have 1 event
	var eventCount int
	err := conn.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM events").Scan(&eventCount)
	require.NoError(t, err)
	require.Equal(t, 1, eventCount, "Expected 1 event to be created")

	// 2. Check ticket_categories table - should have 4 categories (2 seated + 2 free-standing)
	var categoryCount int
	err = conn.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM ticket_categories").Scan(&categoryCount)
	require.NoError(t, err)
	require.Equal(t, 4, categoryCount, "Expected 4 ticket categories to be created")

	// 3. Check ticket_sales table - should have 2 days as specified in DayCount
	var saleCount int
	err = conn.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM ticket_sales").Scan(&saleCount)
	require.NoError(t, err)
	require.Equal(t, 2, saleCount, "Expected 2 ticket sales (days) to be created")

	// 4. Check ticket_packages table - should have 8 packages (4 categories × 2 days)
	var packageCount int
	err = conn.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM ticket_packages").Scan(&packageCount)
	require.NoError(t, err)
	require.Equal(t, 8, packageCount, "Expected 8 ticket packages to be created (4 categories × 2 days)")

	// 5. Check ticket_areas table - should have 10 areas
	// Calculation:
	// - Seated Categories: (1 area × 1 category + 1 area × 1 category) × 2 days = 4 areas
	// - Free Standing: (1 area × 1 category + 2 areas × 1 category) × 2 days = 6 areas
	var areaCount int
	err = conn.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM ticket_areas").Scan(&areaCount)
	require.NoError(t, err)
	require.Equal(t, 10, areaCount, "Expected 10 ticket areas to be created")

	// 6. Check ticket_seats table - should have the sum of all seats
	// Calculation:
	// - Seated Categories: (100 + 250) × 2 days = 700 seats
	// - Free Standing: (300 + 2×1000) × 2 days = 4600 spots
	// Total: 5300 seats/spots
	var seatCount int
	err = conn.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM ticket_seats").Scan(&seatCount)
	require.NoError(t, err)
	require.Equal(t, 5300, seatCount, "Expected 5300 ticket seats to be created")

	// 7. Verify the correct distribution of seats by area type
	var numberedSeatingCount int
	err = conn.Pool.QueryRow(ctx, `
           SELECT COUNT(*) FROM ticket_seats ts
           JOIN ticket_areas ta ON ts.ticket_area_id = ta.id
           WHERE ta.type = 'numbered-seating'
       `).Scan(&numberedSeatingCount)
	require.NoError(t, err)
	require.Equal(t, 700, numberedSeatingCount, "Expected 700 numbered seats")

	var freeStandingCount int
	err = conn.Pool.QueryRow(ctx, `
           SELECT COUNT(*) FROM ticket_seats ts
           JOIN ticket_areas ta ON ts.ticket_area_id = ta.id
           WHERE ta.type = 'free-standing'
       `).Scan(&freeStandingCount)
	require.NoError(t, err)
	require.Equal(t, 4600, freeStandingCount, "Expected 4600 free-standing spots")

	// 8. Verify all seats are initially available
	var availableSeatsCount int
	err = conn.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM ticket_seats WHERE status = 'available'").Scan(&availableSeatsCount)
	require.NoError(t, err)
	require.Equal(t, 5300, availableSeatsCount, "Expected all 5300 seats to have 'available' status")
}
