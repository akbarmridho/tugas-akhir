package invoice

import (
	"context"
	"tugas-akhir/backend/internal/payments/entity"
)

type InvoiceRepository interface {
	CreateInvoice(ctx context.Context, payload entity.CreateInvoiceDto) (*entity.Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, payload entity.UpdateInvoiceStatusDto) error
}
