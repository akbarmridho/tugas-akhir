package event

import (
	"context"
	"testing"
	"tugas-akhir/backend/infrastructure/memcache"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/seeder"
	"tugas-akhir/backend/pkg/utility"
	test_containers "tugas-akhir/backend/test-containers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPGEventRepository is an integration test for the PGEventRepository
func TestPGEventRepository(t *testing.T) {
	for _, variant := range test_containers.RelationalDBVariants {
		t.Run(string(variant), func(t *testing.T) {
			db := seeder.GetConnAndSchema(t, variant)

			// Setup dependencies
			ctx := context.Background()
			cache, merr := memcache.NewMemcache()

			require.NoError(t, merr)

			// Initialize repository
			repo := NewPGEventRepository(db, cache)

			// Seed test data
			eventID := seedTestData(t, ctx, db)

			t.Run("GetEvents", func(t *testing.T) {
				// Call the method to test
				events, err := repo.GetEvents(ctx)

				// Assertions
				require.NoError(t, err, "GetEvents should not return an error")
				require.NotEmpty(t, events, "Events list should not be empty")

				// Verify the event data
				found := false
				for _, e := range events {
					if e.ID == eventID {
						assert.Equal(t, "Music World Tour", e.Name)
						assert.Equal(t, "Jakarta International Stadium", e.Location)
						found = true
						break
					}
				}
				assert.True(t, found, "Seeded event not found in GetEvents result")

				// Test cache hit
				cachedEvents, err := repo.GetEvents(ctx)
				require.NoError(t, err, "GetEvents with cache should not return an error")
				assert.Equal(t, len(events), len(cachedEvents), "Cached events should have same length")

				t.Log(utility.PrettyPrintJSON(events))
				t.Log(utility.PrettyPrintJSON(cachedEvents))
				//assert.Equal(t, true, false)
			})

			t.Run("GetEvent", func(t *testing.T) {
				// Call the method to test
				dto := entity.GetEventDto{ID: eventID}
				foundEvent, err := repo.GetEvent(ctx, dto)

				// Assertions
				require.NoError(t, err, "GetEvent should not return an error")
				require.NotNil(t, foundEvent, "Event should not be nil")

				// Verify the event data
				assert.Equal(t, eventID, foundEvent.ID)
				assert.Equal(t, "Music World Tour", foundEvent.Name)
				assert.Equal(t, "Jakarta International Stadium", foundEvent.Location)

				// Verify relationships
				require.NotEmpty(t, foundEvent.TicketSales, "Event should have ticket sales")

				// Check first ticket sale
				sale := foundEvent.TicketSales[0]
				assert.NotEmpty(t, sale.Name, "Ticket sale should have a name")
				assert.False(t, sale.SaleBeginAt.IsZero(), "Sale begin time should not be zero")
				assert.False(t, sale.SaleEndAt.IsZero(), "Sale end time should not be zero")

				// Check ticket packages within the sale
				require.NotEmpty(t, sale.TicketPackages, "Ticket sale should have packages")

				// Check a ticket package
				pkg := sale.TicketPackages[0]
				assert.Greater(t, pkg.Price, int32(0), "Package price should be positive")

				// Check ticket category
				assert.NotEmpty(t, pkg.TicketCategory.Name, "Ticket category should have a name")

				// Test cache hit
				cachedEvent, err := repo.GetEvent(ctx, dto)
				require.NoError(t, err, "GetEvent with cache should not return an error")
				assert.Equal(t, foundEvent.ID, cachedEvent.ID, "Cached event should have same ID")

				t.Log(utility.PrettyPrintJSON(foundEvent))
				t.Log(utility.PrettyPrintJSON(cachedEvent))
				//assert.Equal(t, true, false)
			})

			t.Run("GetEvent_NotFound", func(t *testing.T) {
				nonExistentID := int64(9999)
				dto := entity.GetEventDto{ID: nonExistentID}

				// Call the method with non-existent ID
				foundEvent, err := repo.GetEvent(ctx, dto)

				// Assertions
				assert.Error(t, err, "GetEvent with non-existent ID should return an error")
				assert.Nil(t, foundEvent, "Event should be nil for non-existent ID")
				assert.ErrorIs(t, err, entity.EventNotFoundError, "Error should be EventNotFoundError")
			})
		})
	}
}

// seedTestData seeds the database with test data and returns the created event ID
func seedTestData(t *testing.T, ctx context.Context, db *postgres.Postgres) int64 {
	// Use the seeder to create test data
	testSeeder := seeder.NewCaseSeeder(db)

	// Define test data payload
	payload := seeder.SeederPayload{
		DayCount: 1, // Just seed one day for testing
		SeatedCategories: []seeder.CategoryPayload{
			{
				Name:        "VIP",
				Price:       1000000,
				AreaCount:   2,
				SeatPerArea: 10,
			},
		},
		FreeStandingCategories: []seeder.CategoryPayload{
			{
				Name:        "Festival",
				Price:       500000,
				AreaCount:   2,
				SeatPerArea: 20,
			},
		},
	}

	// Execute seeder
	err := testSeeder.Seed(ctx, payload)
	require.NoError(t, err, "Failed to seed test data")

	// Query the database to get the created event ID
	var eventID int64
	err = db.GetExecutor(ctx).QueryRow(ctx, "SELECT id FROM events ORDER BY id DESC LIMIT 1").Scan(&eventID)
	require.NoError(t, err, "Failed to get seeded event ID")

	return eventID
}
