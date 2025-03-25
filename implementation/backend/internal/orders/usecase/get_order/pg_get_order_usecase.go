package get_order

import (
	"context"
	"errors"
	"net/http"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/repository/order"
	myerror "tugas-akhir/backend/pkg/error"
)

type PGGetOrderUsecase struct {
	orderRepository order.OrderRepository
}

func NewPGGetOrderUsecase(
	orderRepository order.OrderRepository,
) *PGGetOrderUsecase {
	return &PGGetOrderUsecase{
		orderRepository: orderRepository,
	}
}

func (u *PGGetOrderUsecase) GetOrder(ctx context.Context, payload entity.GetOrderDto) (*entity.Order, *myerror.HttpError) {
	orderEntity, err := u.orderRepository.GetOrder(ctx, payload)

	if err != nil {
		if errors.Is(err, entity.OrderNotFoundError) {
			return nil, &myerror.HttpError{
				Code:         http.StatusNotFound,
				Message:      err.Error(),
				ErrorContext: err,
			}
		}

		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	return orderEntity, nil
}
