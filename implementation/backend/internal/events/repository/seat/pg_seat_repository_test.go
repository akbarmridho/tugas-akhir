package seat

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/seeder"
	"tugas-akhir/backend/pkg/utility"
	test_containers "tugas-akhir/backend/test-containers"
)

// TestPGEventRepository is an integration test for the PGEventRepository
func TestPGSeatRepository(t *testing.T) {
	for _, variant := range test_containers.RelationalDBVariants {
		t.Run(string(variant), func(t *testing.T) {
			ctx := context.Background()

			db := seeder.GetConnAndSchema(t, variant)
			seeder.SeedSchema(t, ctx, db)

			// Initialize repository
			repo := NewPGSeatRepository(db)

			t.Run("GetSeats", func(t *testing.T) {
				// Call the method to test
				seats, err := repo.GetSeats(ctx, entity.GetSeatsDto{
					TicketAreaID: 1,
				})

				// Assertions
				require.NoError(t, err, "GetSeats should not return an error")
				require.NotEmpty(t, seats, "Seats list should not be empty")

				t.Log(utility.PrettyPrintJSON(seats))
				//assert.Equal(t, true, false)
			})

			t.Run("GetSeats_NotFound", func(t *testing.T) {
				nonExistentID := int64(9999)
				dto := entity.GetSeatsDto{TicketAreaID: nonExistentID}

				// Call the method with non-existent ID
				foundSeats, err := repo.GetSeats(ctx, dto)

				// Assertions
				assert.Error(t, err, "GetSeats with non-existent ID should return an error")
				assert.Nil(t, foundSeats, "Seat should be nil for non-existent ID")
				assert.ErrorIs(t, err, entity.SeatNotFoundError, "Error should be SeatNotFoundError")
			})
		})
	}
}
