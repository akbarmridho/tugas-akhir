package order

import (
	"context"
	"tugas-akhir/backend/internal/orders/entity"
)

type OrderRepository interface {
	PlaceOrder(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, error)
	GetOrder(ctx context.Context, payload entity.GetOrderDto) (*entity.Order, error)
	UpdateOrderStatus(ctx context.Context, payload entity.UpdateOrderStatusDto) error
}
