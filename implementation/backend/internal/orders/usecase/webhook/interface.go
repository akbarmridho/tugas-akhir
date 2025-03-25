package webhook

import (
	"context"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/mock_payment"
)

type WebhookOrderUsecase interface {
	HandleWebhook(ctx context.Context, payload mock_payment.Invoice) *myerror.HttpError
}
