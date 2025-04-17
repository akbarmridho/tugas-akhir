package invoice

import (
	"context"
	"testing"
	"time"
	"tugas-akhir/backend/internal/seeder"
	test_containers "tugas-akhir/backend/test-containers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tugas-akhir/backend/internal/payments/entity"
)

func TestPGInvoiceRepository_CreateInvoice(t *testing.T) {
	db := seeder.GetConnAndSchema(t, test_containers.RelationalDBVariant__Postgres)
	seeder.SeedSchema(t, t.Context(), db)

	// Setup repository
	repo := NewPGInvoiceRepository(db)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		var id int64

		insertErr := db.Pool.QueryRow(ctx, "INSERT INTO orders (status, event_id, ticket_sale_id, ticket_area_id, external_user_id) VALUES ('waiting-for-payment', 1, 1, 1, 'user123') RETURNING ID").Scan(&id)

		require.NoError(t, insertErr)

		// Setup test data

		payload := entity.CreateInvoiceDto{
			Amount:       1000,
			ExternalID:   "ext-123",
			OrderID:      id,
			TicketAreaID: 1,
		}

		// Execute the method
		result, err := repo.CreateInvoice(ctx, payload)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, payload.Amount, result.Amount)
		assert.Equal(t, payload.ExternalID, result.ExternalID)
		assert.Equal(t, payload.OrderID, result.OrderID)
		assert.Equal(t, entity.InvoiceStatus__Pending, result.Status)
		assert.NotZero(t, result.ID)
		assert.NotZero(t, result.CreatedAt)
		assert.NotZero(t, result.UpdatedAt)

		// Verify that the invoice was actually created in the database
		var invoiceEntity entity.Invoice
		err = db.Pool.QueryRow(ctx,
			"SELECT id, status, amount, external_id, order_id, created_at, updated_at FROM invoices WHERE id = $1",
			result.ID).Scan(
			&invoiceEntity.ID,
			&invoiceEntity.Status,
			&invoiceEntity.Amount,
			&invoiceEntity.ExternalID,
			&invoiceEntity.OrderID,
			&invoiceEntity.CreatedAt,
			&invoiceEntity.UpdatedAt,
		)
		assert.NoError(t, err)
		assert.Equal(t, result.ID, invoiceEntity.ID)
	})

	t.Run("failure - invalid foreign key constraint", func(t *testing.T) {
		// Setup test data
		ctx := context.Background()
		payload := entity.CreateInvoiceDto{
			Amount:       1000,
			ExternalID:   "ext-123",
			OrderID:      999, // Non-existent order ID
			TicketAreaID: 1,
		}

		// Execute the method
		result, err := repo.CreateInvoice(ctx, payload)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
		// The error should be a foreign key constraint violation
	})
}

func TestPGInvoiceRepository_UpdateInvoiceStatus(t *testing.T) {
	db := seeder.GetConnAndSchema(t, test_containers.RelationalDBVariant__Postgres)
	seeder.SeedSchema(t, t.Context(), db)

	// Setup repository
	repo := NewPGInvoiceRepository(db)

	// Create an invoice first
	ctx := context.Background()

	var id int64

	insertErr := db.Pool.QueryRow(ctx, "INSERT INTO orders (status, event_id, ticket_sale_id, ticket_area_id, external_user_id) VALUES ('waiting-for-payment', 1, 1, 1, 'user123') RETURNING ID").Scan(&id)

	require.NoError(t, insertErr)

	invoiceEntity, err := repo.CreateInvoice(ctx, entity.CreateInvoiceDto{
		Amount:       1000,
		ExternalID:   "ext-123",
		OrderID:      id,
		TicketAreaID: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, invoiceEntity)

	t.Run("success", func(t *testing.T) {
		// Setup test data
		payload := entity.UpdateInvoiceStatusDto{
			ID:     invoiceEntity.ID,
			Status: entity.InvoiceStatus__Paid,
		}

		// Execute the method
		err := repo.UpdateInvoiceStatus(ctx, payload)

		// Assertions
		assert.NoError(t, err)

		// Verify that the status was actually updated in the database
		var status entity.InvoiceStatus
		var updatedAt time.Time
		err = db.Pool.QueryRow(ctx,
			"SELECT status, updated_at FROM invoices WHERE id = $1",
			invoiceEntity.ID).Scan(&status, &updatedAt)
		assert.NoError(t, err)
		assert.Equal(t, entity.InvoiceStatus__Paid, status)
		assert.True(t, updatedAt.After(invoiceEntity.UpdatedAt) || updatedAt.Equal(invoiceEntity.UpdatedAt))
	})
}
