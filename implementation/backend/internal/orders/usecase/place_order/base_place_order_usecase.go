package place_order

import (
	"context"
	"errors"
	errors2 "github.com/pkg/errors"
	"net/http"
	"strconv"
	"time"
	"tugas-akhir/backend/infrastructure/postgres"
	entity2 "tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/bookings/repository/booking"
	entity3 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/repository/event"
	"tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/repository/order"
	entity4 "tugas-akhir/backend/internal/payments/entity"
	"tugas-akhir/backend/internal/payments/repository/invoice"
	"tugas-akhir/backend/internal/payments/service"
	myerror "tugas-akhir/backend/pkg/error"
)

type BasePlaceOrderUsecase struct {
	eventRepository    event.EventRepository
	orderRepository    order.OrderRepository
	bookingRepository  booking.BookingRepository
	invoiceRepository  invoice.InvoiceRepository
	mockPaymentService *service.MockPaymentService
	db                 *postgres.Postgres
}

func NewBasePlaceOrderUsecase(
	eventRepository event.EventRepository,
	orderRepository order.OrderRepository,
	bookingRepository booking.BookingRepository,
	invoiceRepository invoice.InvoiceRepository,
	mockPaymentService *service.MockPaymentService,
	db *postgres.Postgres,
) *BasePlaceOrderUsecase {
	return &BasePlaceOrderUsecase{
		eventRepository:    eventRepository,
		orderRepository:    orderRepository,
		bookingRepository:  bookingRepository,
		invoiceRepository:  invoiceRepository,
		mockPaymentService: mockPaymentService,
		db:                 db,
	}
}

func (u *BasePlaceOrderUsecase) PlaceOrder(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, *myerror.HttpError) {
	if payload.UserID == nil {
		err := errors2.WithStack(errors2.WithMessage(entity3.InternalOrderConfigurationError, "user id is nil"))
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	tx, err := u.db.Pool.Begin(ctx)

	defer tx.Rollback(ctx)

	if err != nil {
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	ctx = context.WithValue(ctx, postgres.PostgresTransactionContextKey, tx)

	eventEntity, err := u.eventRepository.GetEvent(ctx, entity3.GetEventDto{
		ID: payload.EventID,
	})

	if err != nil {
		if errors.Is(err, entity3.EventNotFoundError) {
			return nil, &myerror.HttpError{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}
		}

		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	// enrich payload data
	var ticketSale *entity3.TicketSale

	for _, sale := range eventEntity.TicketSales {
		if sale.ID == payload.TicketSaleID {
			ticketSale = &sale
		}
	}

	if ticketSale == nil {
		return nil, &myerror.HttpError{
			Code:    http.StatusBadRequest,
			Message: entity.TicketSaleNotFoundError.Error(),
		}
	}

	// check for sale time

	now := time.Now()

	if now.Before(ticketSale.SaleBeginAt) {
		return nil, &myerror.HttpError{
			Code:    http.StatusBadRequest,
			Message: entity.TicketSaleNotStartedError.Error(),
		}
	}

	if now.After(ticketSale.SaleEndAt) {
		return nil, &myerror.HttpError{
			Code:    http.StatusBadRequest,
			Message: entity.TicketSaleEndedError.Error(),
		}
	}

	bookRequest := entity2.BookingRequestDto{
		SeatIDs:       []int64{},
		TicketAreaIDs: []int64{},
	}

	for _, item := range payload.Items {
		if item.TicketSeatID != nil {
			bookRequest.SeatIDs = append(bookRequest.SeatIDs, *item.TicketSeatID)
		} else {
			bookRequest.TicketAreaIDs = append(bookRequest.SeatIDs, item.TicketAreaID)
		}
	}

	ticketSeats, err := u.bookingRepository.Book(ctx, bookRequest)

	if err != nil {
		if errors.Is(err, entity2.LockNotAcquiredError) {
			return nil, &myerror.HttpError{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}
		}

		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	payload.FirstTicketAreaID = &ticketSeats[0].TicketAreaID

	total := int32(0)

	takenSeat := make(map[int64]bool)

	// enrich item data
	for _, item := range payload.Items {
		var seat *entity3.TicketSeat

		for _, s := range ticketSeats {
			if item.TicketSeatID == nil {
				// free seated
				if item.TicketAreaID == s.TicketAreaID {
					_, ok := takenSeat[s.ID]

					if ok {
						continue
					} else {
						seat = &s
						takenSeat[s.ID] = true
						break
					}
				}
			} else {
				if *item.TicketSeatID == s.ID {
					seat = &s
					break
				}
			}
		}

		if seat == nil {
			err := errors2.WithStack(errors2.WithMessage(entity3.InternalOrderConfigurationError, "seat is nil"))
			return nil, &myerror.HttpError{
				Code:         http.StatusInternalServerError,
				Message:      err.Error(),
				ErrorContext: err,
			}
		}

		var priceSet = false

		for _, ticketPackage := range ticketSale.TicketPackages {
			for _, area := range ticketPackage.TicketAreas {
				if area.ID == seat.TicketAreaID {
					if area.Type == entity3.AreaType__NumberedSeating && item.TicketSeatID == nil {
						err := errors2.WithStack(errors2.WithMessage(entity3.PlaceOrderBadRequest, "seat is numbered but no seat id given"))
						return nil, &myerror.HttpError{
							Code:         http.StatusBadRequest,
							Message:      err.Error(),
							ErrorContext: err,
						}
					}

					if area.Type == entity3.AreaType__FreeStanding && item.TicketSeatID != nil {
						err := errors2.WithStack(errors2.WithMessage(entity3.PlaceOrderBadRequest, "seat is free standing but seat id given"))
						return nil, &myerror.HttpError{
							Code:         http.StatusBadRequest,
							Message:      err.Error(),
							ErrorContext: err,
						}
					}

					item.TicketCategoryID = &ticketPackage.TicketCategoryID
					item.Price = &ticketPackage.Price
					total += ticketPackage.Price
					priceSet = true
				}
			}
		}

		item.TicketSeatID = &seat.ID

		if !priceSet {
			err := errors2.WithStack(errors2.WithMessage(entity3.InternalOrderConfigurationError, "cannot find area from given payload"))
			return nil, &myerror.HttpError{
				Code:         http.StatusInternalServerError,
				Message:      err.Error(),
				ErrorContext: err,
			}
		}
	}

	orderEntity, err := u.orderRepository.PlaceOrder(ctx, payload)

	if err != nil {
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	invoiceEntity, err := u.invoiceRepository.CreateInvoice(ctx, entity4.CreateInvoiceDto{
		Amount:     total,
		ExternalID: strconv.FormatInt(orderEntity.ID, 10),
		OrderID:    orderEntity.ID,
	})

	if err != nil {
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	orderEntity.Invoice = invoiceEntity

	if err := tx.Commit(ctx); err != nil {
		return nil, &myerror.HttpError{
			Code:         http.StatusInternalServerError,
			Message:      err.Error(),
			ErrorContext: err,
		}
	}

	return orderEntity, nil
}
