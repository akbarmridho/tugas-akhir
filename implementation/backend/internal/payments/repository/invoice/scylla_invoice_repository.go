package invoice

import (
	"context"
	"github.com/pkg/errors"
	"time"
	"tugas-akhir/backend/infrastructure/scylla"
	"tugas-akhir/backend/internal/idgen"
	"tugas-akhir/backend/internal/payments/entity"
)

type ScyllaInvoiceRepository struct {
	scylla *scylla.Scylla
	idgen  *idgen.Idgen
}

func NewScyllaInvoiceRepository(scylla *scylla.Scylla, idgen *idgen.Idgen) *ScyllaInvoiceRepository {
	return &ScyllaInvoiceRepository{
		scylla: scylla,
		idgen:  idgen,
	}
}

func (r *ScyllaInvoiceRepository) CreateInvoice(ctx context.Context, payload entity.CreateInvoiceDto) (*entity.Invoice, error) {
	// Generate ID using Sonyflake
	invoiceID, err := r.idgen.Next()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate invoice ID")
	}

	now := time.Now()

	// Update the order with invoice information
	if err := r.scylla.Session.Query(`
		UPDATE ticket_system.orders 
		SET invoice_id = ?, invoice_status = ?, invoice_amount = ?, 
		    invoice_external_id = ?, invoice_created_at = ?, invoice_updated_at = ? 
		WHERE id = ?`,
		invoiceID, string(entity.InvoiceStatus__Pending), payload.Amount, payload.ExternalID,
		now, now, payload.OrderID,
	).WithContext(ctx).Exec(); err != nil {
		return nil, errors.Wrap(err, "failed to create invoice")
	}

	// Create invoice object to return
	invoice := entity.Invoice{
		ID:         invoiceID,
		Status:     entity.InvoiceStatus__Pending,
		Amount:     payload.Amount,
		ExternalID: payload.ExternalID,
		OrderID:    payload.OrderID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return &invoice, nil
}

func (r *ScyllaInvoiceRepository) UpdateInvoiceStatus(ctx context.Context, payload entity.UpdateInvoiceStatusDto) error {
	now := time.Now()

	// Update invoice fields in the orders table
	if err := r.scylla.Session.Query(`
		UPDATE ticket_system.orders 
		SET invoice_status = ?, invoice_updated_at = ? 
		WHERE id = ?`,
		payload.Status, now, payload.ID,
	).WithContext(ctx).Exec(); err != nil {
		return errors.Wrap(err, "failed to update invoice status")
	}

	return nil
}
