package get_order

import (
	"context"
	"tugas-akhir/backend/internal/orders/entity"
	myerror "tugas-akhir/backend/pkg/error"
)

type GetOrderUsecase interface {
	GetOrder(ctx context.Context, payload entity.GetOrderDto) (*entity.Order, *myerror.HttpError)
}
