package service

import (
	"context"
	"time"
	"tugas-akhir/backend/pkg/mock_payment"
)

type PaymentGatewayMock struct {
}

func NewPaymentGatewayMock() (*PaymentGatewayMock, error) {
	return &PaymentGatewayMock{}, nil
}

func (s *PaymentGatewayMock) GenerateInvoice(ctx context.Context, payload mock_payment.CreateInvoiceRequest) (*mock_payment.Invoice, error) {
	now := time.Now().String()
	exp := time.Now().Add(30 * time.Minute).String()
	return &mock_payment.Invoice{
		Id:         "1",
		Amount:     payload.Amount,
		ExternalId: payload.ExternalId,
		CreatedAt:  *mock_payment.NewNullableString(&now),
		ExpiredAt:  *mock_payment.NewNullableString(&exp),
		Status:     "pending",
	}, nil
}
