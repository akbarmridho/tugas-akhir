package payments

import (
	"go.uber.org/fx"
	"tugas-akhir/backend/internal/payments/repository/invoice"
	"tugas-akhir/backend/internal/payments/service"
)

var BaseModule = fx.Options(
	fx.Provide(fx.Annotate(service.NewMockPaymentService, fx.As(new(service.MockPayment)))),
	fx.Provide(fx.Annotate(invoice.NewPGInvoiceRepository, fx.As(new(invoice.InvoiceRepository)))),
)
