package booked_seats

import (
	"context"
	"testing"
	"tugas-akhir/backend/internal/bookings/service"
	entity3 "tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/seeder"
	test_containers "tugas-akhir/backend/test-containers"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/bookings/entity"

	entity2 "tugas-akhir/backend/internal/events/entity"
)

func TestPGBookedSeatRepository_Integration(t *testing.T) {
	for _, variant := range test_containers.RelationalDBVariants {
		t.Run(string(variant), func(t *testing.T) {
			ctx := context.Background()

			// Setup test database
			db := seeder.GetConnAndSchema(t, variant)
			seeder.SeedSchema(t, ctx, db)

			// Create test data
			seedData := seedTestData(t, ctx, db)

			// Initialize repository with mock serial number generator
			repo := NewPGBookedSeatRepository(db, service.NewSerialNumberGenerator())

			// Test PublishIssuedTickets
			t.Run("PublishIssuedTickets", func(t *testing.T) {
				payload := createPublishTicketPayload(t, db, seedData)
				err := repo.PublishIssuedTickets(ctx, payload)
				require.NoError(t, err)

				// Verify tickets were created
				var issuedTickets []entity.IssuedTicket
				query := `SELECT id, serial_number, holder_name, name, description, ticket_seat_id, order_id, order_item_id, created_at, updated_at FROM issued_tickets WHERE order_id = $1`
				err = pgxscan.Select(ctx, db.GetExecutor(ctx), &issuedTickets, query, payload.Items[0].OrderID)
				require.NoError(t, err)
				assert.Equal(t, len(payload.Items), len(issuedTickets))
			})

			// Test GetIssuedTickets
			t.Run("GetIssuedTickets", func(t *testing.T) {
				// Use the orderID from our seed data
				getTicketPayload := entity.GetIssuedTicketDto{
					OrderID:      seedData.orderID,
					TicketAreaID: seedData.ticketAreaID,
					UserID:       &seedData.externalUserID,
				}

				tickets, err := repo.GetIssuedTickets(ctx, getTicketPayload)
				require.NoError(t, err)
				assert.NotEmpty(t, tickets)

				// Verify ticket details
				for _, ticket := range tickets {
					assert.NotEmpty(t, ticket.SerialNumber)
					assert.Equal(t, seedData.orderID, ticket.OrderID)
					assert.NotNil(t, ticket.TicketSeat)
				}
			})

			// Test GetIssuedTickets with invalid user
			t.Run("GetIssuedTickets_InvalidUser", func(t *testing.T) {
				invalidUser := "invalid-user-id"

				getTicketPayload := entity.GetIssuedTicketDto{
					OrderID:      seedData.orderID,
					TicketAreaID: seedData.ticketAreaID,
					UserID:       &invalidUser,
				}

				_, err := repo.GetIssuedTickets(ctx, getTicketPayload)
				assert.ErrorIs(t, err, entity.IssuedTicketNotFoundError)
			})

			// Test IterSeats
			t.Run("IterSeats", func(t *testing.T) {
				seats, iter, err := repo.IterSeats(ctx)
				require.NoError(t, err)
				require.NotNil(t, iter)

				// Just check that seats array is of expected size (100 as defined in repository)
				assert.Len(t, seats, 100)

				// Verify we can iterate through cursor
				hasNext := iter.Next(ctx)
				assert.True(t, hasNext)
			})
		})
	}
}

// TestData holds references to created test data
type TestData struct {
	eventID          int64
	ticketCategoryID int64
	ticketSaleID     int64
	ticketPackageID  int64
	ticketAreaID     int64
	ticketSeatIDs    []int64
	orderID          int64
	orderItemIDs     []int64
	externalUserID   string
	eventName        string
	ticketSaleName   string
	categoryName     string
}

// seedTestData creates all necessary test data for the test
func seedTestData(t *testing.T, ctx context.Context, db *postgres.Postgres) TestData {
	result := TestData{
		externalUserID: "user123",
		eventName:      "",
		ticketSaleName: "",
		categoryName:   "",
	}

	// Query existing event (you mentioned it's already seeded)
	var events []struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}
	err := pgxscan.Select(ctx, db.GetExecutor(ctx), &events, `SELECT id, name FROM events LIMIT 1`)
	require.NoError(t, err)
	require.NotEmpty(t, events, "No events found in the database. Make sure the event seeder has run.")

	result.eventID = events[0].ID
	result.eventName = events[0].Name

	// Query existing ticket sale
	var ticketSales []struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}
	err = pgxscan.Select(ctx, db.GetExecutor(ctx), &ticketSales,
		`SELECT id, name FROM ticket_sales WHERE event_id = $1 LIMIT 1`, result.eventID)

	require.NoError(t, err)

	result.ticketSaleID = ticketSales[0].ID
	result.ticketSaleName = ticketSales[0].Name

	var ticketPackages []struct {
		ID           int64
		CategoryID   int64
		CategoryName string
	}

	err = pgxscan.Select(ctx, db.GetExecutor(ctx), &ticketPackages,
		`SELECT tp.id as id, tc.id as category_id, tc.name as category_name FROM ticket_packages tp
INNER JOIN ticket_categories tc ON tc.id = tp.ticket_category_id
WHERE tc.event_id = $1
LIMIT 1
`, result.eventID)

	require.NoError(t, err)

	result.ticketCategoryID = ticketPackages[0].CategoryID
	result.categoryName = ticketPackages[0].CategoryName
	result.ticketPackageID = ticketPackages[0].ID

	// ticket area and seat
	var ticketAreas []struct {
		ID int64
	}

	err = pgxscan.Select(ctx, db.GetExecutor(ctx), &ticketAreas,
		`SELECT id FROM ticket_areas WHERE ticket_package_id = $1 LIMIT 1
`, result.ticketPackageID)

	require.NoError(t, err)

	result.ticketAreaID = ticketAreas[0].ID

	var ticketSeats []struct {
		ID int64
	}

	err = pgxscan.Select(ctx, db.GetExecutor(ctx), &ticketSeats,
		`SELECT id FROM ticket_seats WHERE ticket_area_id = $1 LIMIT 2
`, result.ticketAreaID)

	require.NoError(t, err)

	result.ticketSeatIDs = []int64{}

	for _, seat := range ticketSeats {
		result.ticketSeatIDs = append(result.ticketSeatIDs, seat.ID)
	}

	// Create order
	var orderID int64
	err = db.GetExecutor(ctx).QueryRow(ctx,
		`INSERT INTO orders(status, event_id, ticket_sale_id, ticket_area_id, external_user_id) 
		VALUES($1, $2, $3, $4, $5) RETURNING id`,
		"success", result.eventID, result.ticketSaleID, result.ticketAreaID, result.externalUserID).Scan(&orderID)
	require.NoError(t, err)
	result.orderID = orderID

	// Create order items
	result.orderItemIDs = make([]int64, 2)
	for i := 0; i < 2; i++ {
		var orderItemID int64
		err = db.GetExecutor(ctx).QueryRow(ctx,
			`INSERT INTO order_items(customer_name, customer_email, price, order_id, ticket_category_id, ticket_seat_id, ticket_area_id) 
			VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			"Customer "+string('A'+rune(i)),
			"customer"+string('a'+rune(i))+"@example.com",
			100000,
			result.orderID,
			result.ticketCategoryID,
			result.ticketSeatIDs[i],
			result.ticketAreaID).Scan(&orderItemID)
		require.NoError(t, err)
		result.orderItemIDs[i] = orderItemID

		// Update seat status to sold
		_, err = db.GetExecutor(ctx).Exec(ctx,
			`UPDATE ticket_seats SET status = 'sold' WHERE id = $1 AND ticket_area_id = $2`,
			result.ticketSeatIDs[i], result.ticketAreaID)
		require.NoError(t, err)
	}

	return result
}

// createPublishTicketPayload creates the payload for PublishIssuedTickets
func createPublishTicketPayload(t *testing.T, db *postgres.Postgres, data TestData) entity.PublishIssuedTicketDto {
	// Query order items to get accurate data
	type OrderItemData struct {
		ID            int64
		CustomerName  string
		CustomerEmail string
		TicketSeatID  int64
	}

	var orderItems []OrderItemData
	query := `SELECT id, customer_name, customer_email, ticket_seat_id 
			  FROM order_items 
			  WHERE order_id = $1`

	err := pgxscan.Select(context.Background(),
		db.GetExecutor(context.Background()),
		&orderItems, query, data.orderID)
	require.NoError(t, err)
	require.NotEmpty(t, orderItems)

	// Query seat info
	type SeatInfo struct {
		ID         int64
		SeatNumber string
		SeatType   string
	}

	seatInfos := make([]entity.SeatInfoDto, len(orderItems))
	for i, item := range orderItems {
		var info SeatInfo
		query := `SELECT ts.id, ts.seat_number, ta.type as seat_type
				  FROM ticket_seats ts
				  JOIN ticket_areas ta ON ts.ticket_area_id = ta.id
				  WHERE ts.id = $1 AND ts.ticket_area_id = $2`

		err := pgxscan.Get(context.Background(),
			db.GetExecutor(context.Background()),
			&info, query, item.TicketSeatID, data.ticketAreaID)
		require.NoError(t, err)

		seatInfos[i] = entity.SeatInfoDto{
			SeatType:     entity2.AreaType(info.SeatType),
			SeatNumber:   info.SeatNumber,
			CategoryName: data.categoryName,
		}
	}

	// Create items for the payload
	items := make([]entity3.OrderItem, len(orderItems))
	for i, item := range orderItems {
		items[i] = entity3.OrderItem{
			ID:           item.ID,
			OrderID:      data.orderID,
			CustomerName: item.CustomerName,
			TicketSeatID: item.TicketSeatID,
		}
	}

	return entity.PublishIssuedTicketDto{
		EventName:      data.eventName,
		TicketSaleName: data.ticketSaleName,
		TicketAreaID:   data.ticketAreaID,
		Items:          items,
		SeatInfos:      seatInfos,
	}
}
