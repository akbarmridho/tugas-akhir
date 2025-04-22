package booking

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/seeder"
	test_containers "tugas-akhir/backend/test-containers"
)

type SeatStructID struct {
	ID int64
}

func TestPGBookingRepository_Book(t *testing.T) {
	for _, variant := range test_containers.RelationalDBVariants {
		t.Run(string(variant), func(t *testing.T) {
			db := seeder.GetConnAndSchema(t, test_containers.RelationalDBVariant__Postgres)
			seeder.SeedSchema(t, t.Context(), db)

			repo := NewPGBookingRepository(db)

			ctx := context.Background()

			t.Run("Book numbered seats successfully", func(t *testing.T) {
				// Start a transaction for this test
				tx, err := db.Pool.Begin(ctx)
				require.NoError(t, err)
				defer tx.Rollback(ctx) // Always rollback after test

				// Create a new context with the transaction
				txCtx := context.WithValue(ctx, postgres.PostgresTransactionContextKey, tx)

				var ids []SeatStructID

				queryErr := pgxscan.Select(ctx, tx, &ids, `
SELECT ts.id as id
FROM ticket_seats ts
INNER JOIN ticket_areas ta
ON ts.ticket_area_id = ta.id
WHERE ta.type = 'numbered-seating'
ORDER BY ts.id
LIMIT 2
`)

				require.NoError(t, queryErr)
				require.Len(t, ids, 2)

				// Arrange
				payload := entity.BookingRequestDto{
					SeatIDs: []int64{},
				}

				for _, id := range ids {
					payload.SeatIDs = append(payload.SeatIDs, id.ID)
				}

				// Act
				seats, err := repo.Book(txCtx, payload)

				// Assert
				require.NoError(t, err)
				assert.Len(t, seats, len(ids))

				// Verify seats were returned correctly
				seatIDs := []int64{seats[0].ID, seats[1].ID}
				assert.Contains(t, seatIDs, ids[0].ID)
				assert.Contains(t, seatIDs, ids[1].ID)

				// Verify seats are now on-hold
				checkSeatsStatus(t, txCtx, db, []int64{1, 2}, "on-hold")
			})

			t.Run("Book free-standing seats successfully", func(t *testing.T) {
				// Start a transaction for this test
				tx, err := db.Pool.Begin(ctx)
				require.NoError(t, err)
				defer tx.Rollback(ctx) // Always rollback after test

				// Create a new context with the transaction
				txCtx := context.WithValue(ctx, postgres.PostgresTransactionContextKey, tx)

				var ids []SeatStructID

				queryErr := pgxscan.Select(ctx, tx, &ids, `
SELECT id
FROM ticket_areas
WHERE type = 'free-standing'
ORDER BY id
LIMIT 1
`)

				require.NoError(t, queryErr)
				require.Len(t, ids, 1)

				// Arrange
				payload := entity.BookingRequestDto{
					TicketAreaIDs: []int64{ids[0].ID, ids[0].ID}, // Request 2 seats from free-standing area
				}

				// Act
				seats, err := repo.Book(txCtx, payload)

				// Assert
				require.NoError(t, err)
				assert.Len(t, seats, 2)

				// All seats should be from same area id
				for _, seat := range seats {
					assert.Equal(t, ids[0].ID, seat.TicketAreaID)
				}

				// Get the seat IDs that were allocated
				seatIDs := make([]int64, len(seats))
				for i, seat := range seats {
					seatIDs[i] = seat.ID
				}

				// Verify seats are now on-hold
				checkSeatsStatus(t, txCtx, db, seatIDs, "on-hold")
			})

			t.Run("Concurrent booking should cause lock contention", func(t *testing.T) {
				// Start first transaction
				tx1, err := db.Pool.Begin(ctx)
				require.NoError(t, err)
				defer tx1.Rollback(ctx)
				txCtx1 := context.WithValue(ctx, postgres.PostgresTransactionContextKey, tx1)

				var ids []SeatStructID

				queryErr := pgxscan.Select(ctx, tx1, &ids, `
SELECT ts.id as id
FROM ticket_seats ts
INNER JOIN ticket_areas ta
ON ts.ticket_area_id = ta.id
WHERE ta.type = 'numbered-seating'
ORDER BY ts.id DESC
LIMIT 2
`)

				require.NoError(t, queryErr)
				require.Len(t, ids, 2)

				// Book seats in first transaction
				payload := entity.BookingRequestDto{
					SeatIDs: []int64{},
				}

				for _, id := range ids {
					payload.SeatIDs = append(payload.SeatIDs, id.ID)
				}

				seats1, err := repo.Book(txCtx1, payload)
				require.NoError(t, err)
				assert.Len(t, seats1, 2)

				// Start second transaction attempting to book same seats
				tx2, err := db.Pool.Begin(ctx)
				require.NoError(t, err)
				defer tx2.Rollback(ctx)
				txCtx2 := context.WithValue(ctx, postgres.PostgresTransactionContextKey, tx2)

				// Try to book the same seats in second transaction - should fail due to FOR UPDATE NOWAIT
				seats2, err := repo.Book(txCtx2, payload)

				// Should get a lock error
				require.Error(t, err)
				assert.Nil(t, seats2)
				assert.ErrorIs(t, err, entity.LockNotAcquiredError)

				checkSeatsStatus(t, txCtx1, db, payload.SeatIDs, "on-hold")

				// Commit the first transaction
				err = tx1.Commit(ctx)
				require.NoError(t, err)
			})

			t.Run("Book with non-existent seat IDs should fail", func(t *testing.T) {
				// Start a transaction for this test
				tx, err := db.Pool.Begin(ctx)
				require.NoError(t, err)
				defer tx.Rollback(ctx) // Always rollback after test

				// Create a new context with the transaction
				txCtx := context.WithValue(ctx, postgres.PostgresTransactionContextKey, tx)

				// Arrange
				payload := entity.BookingRequestDto{
					SeatIDs: []int64{10000000}, // Non-existent seat ID
				}

				// Act
				seats, err := repo.Book(txCtx, payload)

				// Assert
				require.Error(t, err)
				assert.Nil(t, seats)
				assert.Contains(t, err.Error(), "the result data length does not match")
			})

			t.Run("Book when free-standing area has insufficient seats", func(t *testing.T) {
				// Start a transaction for this test
				tx, err := db.Pool.Begin(ctx)
				require.NoError(t, err)
				defer tx.Rollback(ctx) // Always rollback after test

				// Create a new context with the transaction
				txCtx := context.WithValue(ctx, postgres.PostgresTransactionContextKey, tx)

				var ticketAreaID int64

				insertErr := tx.QueryRow(ctx, "INSERT INTO ticket_areas (type, ticket_package_id) VALUES ('free-standing', 1) RETURNING id").Scan(&ticketAreaID)

				require.NoError(t, insertErr)

				// Arrange - Request more seats than available in the area
				payload := entity.BookingRequestDto{
					TicketAreaIDs: []int64{ticketAreaID, ticketAreaID, ticketAreaID, ticketAreaID}, // Request 4 seats where no seat
				}

				// Act
				seats, err := repo.Book(txCtx, payload)

				// Assert
				require.Error(t, err)
				assert.Nil(t, seats)
				assert.Contains(t, err.Error(), "cannot acquire locks for the given seats")
			})
		})
	}
}

// Helper function to check if seats have the expected status
func checkSeatsStatus(t *testing.T, ctx context.Context, db *postgres.Postgres, seatIDs []int64, expectedStatus string) {
	query := `
		SELECT id, status 
		FROM ticket_seats 
		WHERE id = ANY($1)
	`

	rows, err := db.GetExecutor(ctx).Query(ctx, query, seatIDs)
	require.NoError(t, err)
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int64
		var status string
		err := rows.Scan(&id, &status)
		require.NoError(t, err)
		assert.Equal(t, expectedStatus, status)
		count++
	}

	assert.Equal(t, len(seatIDs), count, "Should find all requested seats")
}
