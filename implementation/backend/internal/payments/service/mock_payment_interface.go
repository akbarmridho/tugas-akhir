package service

import (
	"context"
	"tugas-akhir/backend/pkg/mock_payment"
)

type MockPayment interface {
	GenerateInvoice(ctx context.Context, payload mock_payment.CreateInvoiceRequest) (*mock_payment.Invoice, error)
}
