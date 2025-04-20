package webhook

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"tugas-akhir/backend/infrastructure/postgres"
	entity3 "tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/bookings/repository/booked_seats"
	"tugas-akhir/backend/internal/bookings/repository/booking"
	entity4 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/repository/event"
	"tugas-akhir/backend/internal/events/service/redis_availability_seeder"
	entity2 "tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/repository/order"
	"tugas-akhir/backend/internal/orders/service/early_dropper"
	"tugas-akhir/backend/internal/payments/entity"
	"tugas-akhir/backend/internal/payments/repository/invoice"
	myerror "tugas-akhir/backend/pkg/error"
	"tugas-akhir/backend/pkg/mock_payment"
)

type PGWebhookUsecase struct {
	orderRepository         order.OrderRepository
	invoiceRepository       invoice.InvoiceRepository
	eventRepository         event.EventRepository
	bookeadSeatRepository   booked_seats.BookedSeatRepository
	bookingRepository       booking.BookingRepository
	redisAvailabilitySeeder *redis_availability_seeder.RedisAvailabilitySeeder
	db                      *postgres.Postgres
	earlyDropper            *early_dropper.EarlyDropper
}

func NewPGWebhookUsecase(
	orderRepository order.OrderRepository,
	invoiceRepository invoice.InvoiceRepository,
	bookeadSeatRepository booked_seats.BookedSeatRepository,
	bookingRepository booking.BookingRepository,
	eventRepository event.EventRepository,
	redisAvailabilitySeeder *redis_availability_seeder.RedisAvailabilitySeeder,
	db *postgres.Postgres,
) *PGWebhookUsecase {
	return &PGWebhookUsecase{
		orderRepository:         orderRepository,
		invoiceRepository:       invoiceRepository,
		bookeadSeatRepository:   bookeadSeatRepository,
		eventRepository:         eventRepository,
		db:                      db,
		bookingRepository:       bookingRepository,
		redisAvailabilitySeeder: redisAvailabilitySeeder,
	}
}

func NewFCWebhookUsecase(
	orderRepository order.OrderRepository,
	invoiceRepository invoice.InvoiceRepository,
	bookeadSeatRepository booked_seats.BookedSeatRepository,
	bookingRepository booking.BookingRepository,
	eventRepository event.EventRepository,
	redisAvailabilitySeeder *redis_availability_seeder.RedisAvailabilitySeeder,
	db *postgres.Postgres,
	earlyDropper *early_dropper.EarlyDropper,
) *PGWebhookUsecase {
	return &PGWebhookUsecase{
		orderRepository:         orderRepository,
		invoiceRepository:       invoiceRepository,
		bookeadSeatRepository:   bookeadSeatRepository,
		eventRepository:         eventRepository,
		db:                      db,
		bookingRepository:       bookingRepository,
		redisAvailabilitySeeder: redisAvailabilitySeeder,
		earlyDropper:            earlyDropper,
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

	seatIDs := make([]int64, 0)

	for _, item := range orderEntity.Items {
		seatIDs = append(seatIDs, item.TicketSeatID)
	}

	if shouldPublish {
		if u.earlyDropper != nil {
			err = u.earlyDropper.FinalizeLock(ctx, orderEntity.Items, entity4.SeatStatus__Sold)

			if err != nil {
				return &myerror.HttpError{
					Code:         http.StatusInternalServerError,
					Message:      err.Error(),
					ErrorContext: err,
				}
			}
		}

		err = u.bookingRepository.UpdateSeatStatus(ctx, entity3.UpdateSeatStatusDto{
			SeatIDs: seatIDs,
			Status:  entity4.SeatStatus__Sold,
		})

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

		err = u.bookeadSeatRepository.PublishIssuedTickets(ctx, entity3.PublishIssuedTicketDto{
			EventName:      fmt.Sprintf("%s - %s", orderEntity.Event.Name, orderEntity.Event.Location),
			TicketSaleName: orderEntity.TicketSale.Name,
			SeatInfos:      seatInfo,
			Items:          orderEntity.Items,
			TicketAreaID:   orderEntity.TicketAreaID,
		})

		if err != nil {
			return &myerror.HttpError{
				Code:         http.StatusInternalServerError,
				Message:      err.Error(),
				ErrorContext: err,
			}
		}
	} else {
		err = u.bookingRepository.UpdateSeatStatus(ctx, entity3.UpdateSeatStatusDto{
			SeatIDs: seatIDs,
			Status:  entity4.SeatStatus__Available,
		})

		if err != nil {
			return &myerror.HttpError{
				Code:         http.StatusInternalServerError,
				Message:      err.Error(),
				ErrorContext: err,
			}
		}

		revertedAvailability := make([]entity4.AreaAvailability, 0)

		for _, item := range orderEntity.Items {
			revertedAvailability = append(revertedAvailability, entity4.AreaAvailability{
				TicketAreaID:    item.TicketSeat.TicketAreaID,
				TicketPackageID: item.TicketSeat.TicketArea.TicketPackageID,
				TicketSaleID:    orderEntity.TicketSaleID,
			})
		}

		if u.earlyDropper != nil {
			err = u.earlyDropper.FinalizeLock(ctx, orderEntity.Items, entity4.SeatStatus__Available)

			if err != nil {
				return &myerror.HttpError{
					Code:         http.StatusInternalServerError,
					Message:      err.Error(),
					ErrorContext: err,
				}
			}
		}

		err = u.redisAvailabilitySeeder.RevertAvailability(ctx, revertedAvailability)

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
