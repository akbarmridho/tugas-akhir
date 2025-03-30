package get_order

import (
	"context"
	entity2 "tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/orders/entity"
	myerror "tugas-akhir/backend/pkg/error"
)

type GetOrderUsecase interface {
	GetOrder(ctx context.Context, payload entity.GetOrderDto) (*entity.Order, *myerror.HttpError)
	GetIssuedTicket(ctx context.Context, payload entity2.GetIssuedTicketDto) ([]entity2.IssuedTicket, *myerror.HttpError)
}
