package place_order

import (
	"context"
	"tugas-akhir/backend/internal/orders/entity"
	myerror "tugas-akhir/backend/pkg/error"
)

type PlaceOrderUsecase interface {
	PlaceOrder(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError)
}
