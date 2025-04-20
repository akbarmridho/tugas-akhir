package get_order

import (
	"context"
	"errors"
	"net/http"
	entity2 "tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/bookings/repository/booked_seats"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/repository/order"
	myerror "tugas-akhir/backend/pkg/error"
)

type PGGetOrderUsecase struct {
	orderRepository order.OrderRepository
	seatRepository  booked_seats.BookedSeatRepository
}

func NewPGGetOrderUsecase(
	orderRepository order.OrderRepository,
	seatRepository booked_seats.BookedSeatRepository,
) *PGGetOrderUsecase {
	return &PGGetOrderUsecase{
		orderRepository: orderRepository,
		seatRepository:  seatRepository,
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

func (u *PGGetOrderUsecase) GetIssuedTicket(ctx context.Context, payload entity2.GetIssuedTicketDto) ([]entity2.IssuedTicket, *myerror.HttpError) {
	issuedTickets, err := u.seatRepository.GetIssuedTickets(ctx, payload)

	if err != nil {
		if errors.Is(err, entity2.IssuedTicketNotFoundError) {
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

	return issuedTickets, nil
}
