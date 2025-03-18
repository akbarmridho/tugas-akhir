package order

import (
	"context"
	"tugas-akhir/backend/internal/orders/entity"
)

type OrderRepository interface {
	PlaceOrder(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, error)
}
