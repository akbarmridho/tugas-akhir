package invoice

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/pkg/errors"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/payments/entity"
)

type PGInvoiceRepository struct {
	db *postgres.Postgres
}

func NewPGInvoiceRepository(db *postgres.Postgres) *PGInvoiceRepository {
	return &PGInvoiceRepository{
		db: db,
	}
}

func (r *PGInvoiceRepository) CreateInvoice(ctx context.Context, payload entity.CreateInvoiceDto) (*entity.Invoice, error) {
	query := `
	INSERT INTO invoices(status, amount, external_id, order_id, ticket_area_id)
	VALUES ('pending', $1, $2, $3, $4)
    `

	var invoice entity.Invoice

	err := pgxscan.Get(ctx, r.db.GetExecutor(ctx), &invoice, query, payload.Amount, payload.ExternalID, payload.OrderID, payload.TicketAreaID)

	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, errors.WithMessage(entity.CreateInvoiceInternalError, "no invoice returned")
		}

		return nil, err
	}

	return &invoice, nil
}

func (r *PGInvoiceRepository) UpdateInvoiceStatus(ctx context.Context, payload entity.UpdateInvoiceStatusDto) error {
	query := `
	UPDATE invoices
	SET status = $1, updated_at = now()
	WHERE id = $2
    `

	_, err := r.db.GetExecutor(ctx).Exec(ctx, query, payload.Status, payload.ID)

	return err
}
