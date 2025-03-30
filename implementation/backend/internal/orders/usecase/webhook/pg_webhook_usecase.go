package webhook

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"tugas-akhir/backend/infrastructure/postgres"
	entity3 "tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/bookings/repository/booking"
	"tugas-akhir/backend/internal/events/repository/event"
	entity2 "tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/repository/order"
	"tugas-akhir/backend/internal/payments/entity"
	"tugas-akhir/backend/internal/payments/repository/invoice"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/mock_payment"
)

type PGWebhookUsecase struct {
	orderRepository   order.OrderRepository
	invoiceRepository invoice.InvoiceRepository
	eventRepository   event.EventRepository
	bookingRepository booking.BookingRepository
	db                *postgres.Postgres
}

func NewPGWebhookUsecase(
	orderRepository order.OrderRepository,
	invoiceRepository invoice.InvoiceRepository,
	bookingRepository booking.BookingRepository,
	eventRepository event.EventRepository,
	db *postgres.Postgres,
) *PGWebhookUsecase {
	return &PGWebhookUsecase{
		orderRepository:   orderRepository,
		invoiceRepository: invoiceRepository,
		bookingRepository: bookingRepository,
		eventRepository:   eventRepository,
		db:                db,
	}
}

func (u *PGWebhookUsecase) HandleWebhook(ctx context.Context, payload mock_payment.Invoice) *myerror.HttpError {
	tx, err := u.db.Pool.Begin(ctx)

	defer tx.Rollback(ctx)

	if err != nil {
		return &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	ctx = context.WithValue(ctx, postgres.PostgresTransactionContextKey, tx)

	orderID, err := strconv.ParseInt(payload.ExternalId, 10, 64)

	if err != nil {
		err := errors.WithStack(errors.WithMessage(entity2.WebhookInternalError, fmt.Sprintf("cannot parse order id %s", payload.ExternalId)))
		return &myerror.HttpError{
			Code:         http.StatusBadRequest,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	orderEntity, err := u.orderRepository.GetOrder(ctx, entity2.GetOrderDto{
		OrderID:      orderID,
		BypassUserID: true,
	})

	if err != nil {
		if errors.Is(err, entity2.OrderNotFoundError) {
			return &myerror.HttpError{
				Code:         http.StatusNotFound,
				Message:      err.Error(),
				ErrorContext: err,
			}
		}

		return &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	updateInvoice := entity.UpdateInvoiceStatusDto{
		ID: orderEntity.Invoice.ID,
	}

	updateOrder := entity2.UpdateOrderStatusDto{
		OrderID: orderID,
	}

	shouldPublish := false

	if payload.Status == "expired" {
		failReason := "payment expired"
		updateOrder.FailReason = &failReason
		updateOrder.Status = entity2.OrderStatus__Failed
		updateInvoice.Status = entity.InvoiceStatus__Expired
	} else if payload.Status == "paid" {
		updateOrder.Status = entity2.OrderStatus__Success
		updateInvoice.Status = entity.InvoiceStatus__Paid
		shouldPublish = true
	} else if payload.Status == "failed" {
		failReason := "payment failed"
		updateOrder.FailReason = &failReason
		updateOrder.Status = entity2.OrderStatus__Failed
		updateInvoice.Status = entity.InvoiceStatus__Failed
	} else {
		err := errors.WithStack(errors.WithMessage(entity2.WebhookInternalError, fmt.Sprintf("unexpected status %s", payload.Status)))
		return &myerror.HttpError{
			Code:         http.StatusBadRequest,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	err = u.invoiceRepository.UpdateInvoiceStatus(ctx, updateInvoice)

	if err != nil {
		return &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	err = u.orderRepository.UpdateOrderStatus(ctx, updateOrder)

	if err != nil {
		return &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	seatInfo := make([]entity3.SeatInfoDto, 0)

	for _, item := range orderEntity.Items {
		seatInfo = append(seatInfo, entity3.SeatInfoDto{
			CategoryName: item.TicketCategory.Name,
			SeatType:     item.TicketSeat.TicketArea.Type,
			SeatNumber:   item.TicketSeat.SeatNumber,
		})
	}

	if shouldPublish {
		err = u.bookingRepository.PublishIssuedTickets(ctx, entity3.PublishIssuedTicketDto{
			EventName:      fmt.Sprintf("%s - %s", orderEntity.Event.Name, orderEntity.Event.Location),
			TicketSaleName: orderEntity.TicketSale.Name,
			SeatInfos:      seatInfo,
			Items:          orderEntity.Items,
		})

		if err != nil {
			return &myerror.HttpError{
				Code:         http.StatusInternalServerError,
				Message:      err.Error(),
				ErrorContext: err,
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	return nil
}
