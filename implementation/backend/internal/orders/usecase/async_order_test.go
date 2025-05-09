package usecase

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"
	"tugas-akhir/backend/app/processor"
	"tugas-akhir/backend/app/processor/worker"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/memcache"
	"tugas-akhir/backend/infrastructure/postgres"
	entity4 "tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/bookings/repository/booked_seats"
	"tugas-akhir/backend/internal/bookings/repository/booking"
	service2 "tugas-akhir/backend/internal/bookings/service"
	entity2 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/repository/availability"
	"tugas-akhir/backend/internal/events/repository/event"
	"tugas-akhir/backend/internal/events/repository/seat"
	"tugas-akhir/backend/internal/events/service/redis_availability_seeder"
	entity3 "tugas-akhir/backend/internal/orders/entity"
	"tugas-akhir/backend/internal/orders/repository/order"
	"tugas-akhir/backend/internal/orders/service/early_dropper"
	"tugas-akhir/backend/internal/orders/service/pgp_place_order_connector"
	"tugas-akhir/backend/internal/orders/usecase/get_order"
	"tugas-akhir/backend/internal/orders/usecase/place_order"
	"tugas-akhir/backend/internal/orders/usecase/webhook"
	"tugas-akhir/backend/internal/payments/entity"
	"tugas-akhir/backend/internal/payments/repository/invoice"
	"tugas-akhir/backend/internal/payments/service"
	"tugas-akhir/backend/internal/seeder"
	"tugas-akhir/backend/pkg/mock_payment"
	"tugas-akhir/backend/pkg/utility"
	test_containers "tugas-akhir/backend/test-containers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type asyncontainer struct {
	db                     *postgres.Postgres
	orderRepo              order.OrderRepository
	eventRepo              event.EventRepository
	bookingRepo            booking.BookingRepository
	invoiceRepo            invoice.InvoiceRepository
	bookedSeatRepo         booked_seats.BookedSeatRepository
	seatRepo               seat.SeatRepository
	availabilityRepository availability.AvailabilityRepository
	placeOrderUsecase      place_order.PlaceOrderUsecase
	webhookOrderUsecase    webhook.WebhookOrderUsecase
	getOrderUsecase        get_order.GetOrderUsecase
	asyncProcessor         *processor.Processor
}

func setupAsyncTestEnvironment(t *testing.T, variant test_containers.RelationalDBVariant) asyncontainer {
	ctx := context.Background()
	db := seeder.GetConnAndSchema(t, variant)
	seeder.SeedSchema(t, ctx, db)

	rabbitMQ := test_containers.GetRabbitMQContainer(t)
	rmqHost, err := rabbitMQ.Host(ctx)
	require.NoError(t, err)

	rmqPort, err := rabbitMQ.MappedPort(ctx, rabbitmq.DefaultAMQPPort)
	require.NoError(t, err)

	redisClient := test_containers.GetRedisCluster(t)

	cfg := &config.Config{
		PodName: "test-pod-1",
		AmqpUrl: fmt.Sprintf("amqp://%s:%s@%s:%s/", "guest", "guest", rmqHost, rmqPort.Port()),
	}

	t.Logf("amqp url: %s", cfg.AmqpUrl)

	redisSeeder := redis_availability_seeder.NewRedisAvailabilitySeeder(cfg, redisClient, db)
	cache, merr := memcache.NewMemcache()
	require.NoError(t, merr)
	err = redisSeeder.Run(ctx)
	require.NoError(t, err)

	var orderRepo order.OrderRepository
	var eventRepo event.EventRepository
	var bookingRepo booking.BookingRepository
	var invoiceRepo invoice.InvoiceRepository
	var bookedSeatRepo booked_seats.BookedSeatRepository
	var seatRepo seat.SeatRepository
	var availabilityRepository availability.AvailabilityRepository

	orderRepo = order.NewPGOrderRepository(db)
	eventRepo = event.NewPGEventRepository(db, cache)
	bookingRepo = booking.NewPGBookingRepository(db)
	invoiceRepo = invoice.NewPGInvoiceRepository(db)
	bookedSeatRepo = booked_seats.NewPGBookedSeatRepository(db, service2.NewSerialNumberGenerator())
	seatRepo = seat.NewPGSeatRepository(db)
	availabilityRepository = availability.NewRedisAvailabilityRepository(redisClient, cache)

	var paymentGateway service.PaymentGateway
	paymentGateway, _ = service.NewPaymentGatewayMock()

	// Initialize Usecases
	basePlaceOrder := place_order.NewBasePlaceOrderUsecase(
		eventRepo,
		orderRepo,
		bookingRepo,
		invoiceRepo,
		paymentGateway,
		redisSeeder,
		db,
	)

	connector := pgp_place_order_connector.NewFCPlaceOrderConnector(cfg)
	require.NoError(t, connector.Run(ctx))
	earlyDropper := early_dropper.NewFCEarlyDropper(cfg, redisClient, bookedSeatRepo)
	require.NoError(t, earlyDropper.Run(ctx))

	placeOrderUsecase := place_order.NewFCPlaceOrderUsecase(cfg, connector, earlyDropper)

	webhookUsecase := webhook.NewFCWebhookUsecase(
		orderRepo,
		invoiceRepo,
		bookedSeatRepo,
		bookingRepo,
		eventRepo,
		redisSeeder,
		db,
		earlyDropper,
	)

	getOrderUsecase := get_order.NewPGGetOrderUsecase(
		orderRepo,
		bookedSeatRepo,
	)

	resultPublisher := worker.NewResultPublisher(cfg)
	bookingWorker := worker.NewBookingWorker(basePlaceOrder, resultPublisher)
	asyncProcessor, err := processor.NewProcessor(cfg, bookingWorker)
	require.NoError(t, err)
	require.NoError(t, asyncProcessor.Run(ctx))

	t.Cleanup(func() {
		t.Log(asyncProcessor.Stop())
		t.Log(bookingWorker.Stop())
		t.Log(resultPublisher.Stop())
		t.Log(earlyDropper.Stop())
		t.Log(connector.Stop())
	})

	return asyncontainer{
		db:                     db,
		orderRepo:              orderRepo,
		eventRepo:              eventRepo,
		bookingRepo:            bookingRepo,
		invoiceRepo:            invoiceRepo,
		bookedSeatRepo:         bookedSeatRepo,
		seatRepo:               seatRepo,
		placeOrderUsecase:      placeOrderUsecase,
		webhookOrderUsecase:    webhookUsecase,
		getOrderUsecase:        getOrderUsecase,
		availabilityRepository: availabilityRepository,
		asyncProcessor:         asyncProcessor,
	}
}

func selectAsyncByAreaType(t *testing.T, app asyncontainer, userID string, eventEntity *entity2.Event, areaType entity2.AreaType) (
	*entity2.TicketArea,
	*entity2.TicketPackage,
	*entity2.TicketSale,
	entity3.PlaceOrderDto,
) {
	ctx := t.Context()

	var ticketArea *entity2.TicketArea
	var ticketPackage *entity2.TicketPackage
	ticketSale := eventEntity.TicketSales[0]

mainLoop:
	for _, salePackage := range ticketSale.TicketPackages {
		for _, area := range salePackage.TicketAreas {
			if area.Type == areaType {
				ticketPackage = &salePackage
				ticketArea = &area
				break mainLoop
			}
		}
	}

	require.NotNil(t, ticketArea)
	require.NotNil(t, ticketPackage)

	seats, err := app.seatRepo.GetSeats(ctx, entity2.GetSeatsDto{
		TicketAreaID: ticketArea.ID,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(seats), 2)

	var placeOrderPayload entity3.PlaceOrderDto

	if areaType == entity2.AreaType__NumberedSeating {
		placeOrderPayload = entity3.PlaceOrderDto{
			UserID:       &userID,
			EventID:      eventEntity.ID,
			TicketSaleID: ticketSale.ID,
			TicketAreaID: &ticketArea.ID,
			Items: []entity3.OrderItemDto{
				{
					CustomerName:  "Customer A",
					CustomerEmail: "customer.a@example.com",
					TicketSeatID:  &seats[0].ID,
					TicketAreaID:  ticketArea.ID,
				},
				{
					CustomerName:  "Customer B",
					CustomerEmail: "customer.b@example.com",
					TicketSeatID:  &seats[1].ID,
					TicketAreaID:  ticketArea.ID,
				},
			},
		}
	} else {
		placeOrderPayload = entity3.PlaceOrderDto{
			UserID:       &userID,
			EventID:      eventEntity.ID,
			TicketSaleID: ticketSale.ID,
			TicketAreaID: &ticketArea.ID,
			Items: []entity3.OrderItemDto{
				{
					CustomerName:  "Customer A",
					CustomerEmail: "customer.a@example.com",
					TicketAreaID:  ticketArea.ID,
				},
				{
					CustomerName:  "Customer B",
					CustomerEmail: "customer.b@example.com",
					TicketAreaID:  ticketArea.ID,
				},
			},
		}
	}

	key := RandStringBytes(10)
	placeOrderPayload.IdempotencyKey = &key

	return ticketArea, ticketPackage, &ticketSale, placeOrderPayload
}

// --- Test Cases ---

func TestIntegration_Async_OrderFlow_Success(t *testing.T) {
	for _, variant := range test_containers.RelationalDBVariants {
		t.Run(string(variant), func(t *testing.T) {
			ctx := t.Context()

			app := setupAsyncTestEnvironment(t, variant)
			placeOrderUsecase := app.placeOrderUsecase
			webhookUsecase := app.webhookOrderUsecase
			getOrderUsecase := app.getOrderUsecase

			events, err := app.eventRepo.GetEvents(ctx)
			require.NoError(t, err)

			eventEntity, err := app.eventRepo.GetEvent(ctx, entity2.GetEventDto{
				ID: events[0].ID,
			})
			require.NoError(t, err)

			areaTypes := []entity2.AreaType{entity2.AreaType__NumberedSeating, entity2.AreaType__FreeStanding}

			// schrodinger bug
			// if you don't sleep for enough time the rabbitmq queue will not be ready yet to accept incoming messages
			time.Sleep(10 * time.Second)

			for _, areaType := range areaTypes {
				t.Run(fmt.Sprintf("Area Type %s", string(areaType)), func(t *testing.T) {
					userID := fmt.Sprintf("user-%s", string(areaType))

					ticketArea, _, ticketSale, placeOrderPayload := selectAsyncByAreaType(t, app, userID, eventEntity, areaType)

					availabilities, err := app.availabilityRepository.GetAvailability(ctx, entity2.GetAvailabilityDto{
						TicketSaleID: ticketSale.ID,
					})
					require.NoError(t, err)

					var initialAvailability *entity2.AreaAvailability

					for _, currentAvailability := range availabilities {
						if currentAvailability.TicketAreaID == ticketArea.ID {
							initialAvailability = &currentAvailability
						}
					}

					require.NotNil(t, initialAvailability)

					// --- Act: Place Order ---
					placedOrder, placeErr := placeOrderUsecase.PlaceOrder(ctx, placeOrderPayload)

					// --- Assert: Place Order ---
					if placeErr != nil {
						require.NoError(t, placeErr.ErrorContext, "PlaceOrder should succeed")
					}
					require.Nil(t, placeErr, "PlaceOrder should succeed")

					t.Log(utility.PrettyPrintJSON(placedOrder))

					require.NotNil(t, placedOrder, "Placed order should not be nil")
					require.NotNil(t, placedOrder.ID, "Placed order should have an ID")
					require.Equal(t, entity3.OrderStatus__WaitingForPayment, placedOrder.Status, "Initial order status should be waiting-for-payment")
					require.Len(t, placedOrder.Items, len(placeOrderPayload.Items), "Number of items should match payload")
					require.NotNil(t, placedOrder.Invoice, "Order should have an associated invoice")
					require.NotZero(t, placedOrder.Invoice.Amount, "Invoice amount should be calculated")
					require.Equal(t, "1", placedOrder.Invoice.ExternalID) // the id of external payment gateway

					// --- Act: Handle Webhook (Success) ---
					webhookPayloadSuccess := mock_payment.Invoice{
						ExternalId: strconv.FormatInt(placedOrder.ID, 10),
						Status:     "paid",
						Amount:     float32(placedOrder.Invoice.Amount),
					}
					webhookErrSuccess := webhookUsecase.HandleWebhook(ctx, webhookPayloadSuccess)

					// --- Assert: Handle Webhook (Success) ---
					if webhookErrSuccess != nil {
						require.NoError(t, webhookErrSuccess.ErrorContext, "Handling successful webhook should not return an error")
					}
					require.Nil(t, webhookErrSuccess, "Handling successful webhook should not return an error")

					// --- Act: Get Order (After Success) ---
					getOrderPayload := entity3.GetOrderDto{
						OrderID: placedOrder.ID,
						UserID:  &userID,
					}
					fetchedOrderSuccess, getOrderErrSuccess := getOrderUsecase.GetOrder(ctx, getOrderPayload)

					// --- Assert: Get Order (After Success) ---
					if getOrderErrSuccess != nil {
						require.NoError(t, getOrderErrSuccess.ErrorContext, "GetOrder after success should succeed")
					}
					require.Nil(t, getOrderErrSuccess, "GetOrder after success should succeed")
					require.NotNil(t, fetchedOrderSuccess, "Fetched order after success should not be nil")

					t.Log(utility.PrettyPrintJSON(fetchedOrderSuccess))

					require.Equal(t, placedOrder.ID, fetchedOrderSuccess.ID, "Fetched order ID should match placed order ID")
					require.Equal(t, entity3.OrderStatus__Success, fetchedOrderSuccess.Status, "Order status should be success")
					require.Nil(t, fetchedOrderSuccess.FailReason, "FailReason should be nil for successful order")
					require.NotNil(t, fetchedOrderSuccess.Invoice, "Fetched order should include invoice details")
					require.Equal(t, entity.InvoiceStatus__Paid, fetchedOrderSuccess.Invoice.Status, "Invoice status should be paid")
					require.Len(t, fetchedOrderSuccess.Items, len(placedOrder.Items), "Fetched order should have correct number of items")

					for _, item := range fetchedOrderSuccess.Items {
						require.Equal(t, entity2.SeatStatus__Sold, item.TicketSeat.Status)
					}

					availabilities, err = app.availabilityRepository.GetAvailability(ctx, entity2.GetAvailabilityDto{
						TicketSaleID: ticketSale.ID,
					})
					require.NoError(t, err)

					var afterAvailability *entity2.AreaAvailability

					for _, currentAvailability := range availabilities {
						if currentAvailability.TicketAreaID == ticketArea.ID {
							afterAvailability = &currentAvailability
						}
					}

					require.NotNil(t, afterAvailability)
					require.Equal(t, initialAvailability.AvailableSeats-2, afterAvailability.AvailableSeats)

					// --- Act: Get Issued Tickets ---
					getTicketsPayload := entity4.GetIssuedTicketDto{
						ID:     placedOrder.ID,
						UserID: &userID,
					}
					issuedTickets, getTicketsErr := getOrderUsecase.GetIssuedTicket(ctx, getTicketsPayload)

					// --- Assert: Get Issued Tickets ---
					require.Nil(t, getTicketsErr, "GetIssuedTicket should succeed for a successful order")
					require.NotNil(t, issuedTickets, "Issued tickets slice should not be nil")

					t.Log(utility.PrettyPrintJSON(issuedTickets))

					require.Len(t, issuedTickets, len(placedOrder.Items), "Should receive one ticket per order item")
					assert.Equal(t, placeOrderPayload.Items[0].CustomerName, issuedTickets[0].HolderName)
					assert.Equal(t, placeOrderPayload.Items[1].CustomerName, issuedTickets[1].HolderName)
				})
			}
		})
	}
}

func TestIntegration_Async_OrderFlow_PaymentFailed(t *testing.T) {
	for _, variant := range test_containers.RelationalDBVariants {
		t.Run(string(variant), func(t *testing.T) {
			ctx := t.Context()

			app := setupAsyncTestEnvironment(t, variant)
			placeOrderUsecase := app.placeOrderUsecase
			webhookUsecase := app.webhookOrderUsecase
			getOrderUsecase := app.getOrderUsecase

			events, err := app.eventRepo.GetEvents(ctx)
			require.NoError(t, err)

			eventEntity, err := app.eventRepo.GetEvent(ctx, entity2.GetEventDto{
				ID: events[0].ID,
			})
			require.NoError(t, err)

			// schrodinger bug
			// if you don't sleep for enough time the rabbitmq queue will not be ready yet to accept incoming messages
			time.Sleep(10 * time.Second)

			userID := "user-failpayment"

			ticketArea, _, ticketSale, placeOrderPayload := selectAsyncByAreaType(t, app, userID, eventEntity, entity2.AreaType__NumberedSeating)

			availabilities, err := app.availabilityRepository.GetAvailability(ctx, entity2.GetAvailabilityDto{
				TicketSaleID: ticketSale.ID,
			})
			require.NoError(t, err)

			var initialAvailability *entity2.AreaAvailability

			for _, currentAvailability := range availabilities {
				if currentAvailability.TicketAreaID == ticketArea.ID {
					initialAvailability = &currentAvailability
				}
			}

			require.NotNil(t, initialAvailability)

			// --- Act: Place Order ---
			placedOrder, placeErr := placeOrderUsecase.PlaceOrder(ctx, placeOrderPayload)

			// --- Assert: Place Order ---
			if placeErr != nil {
				require.NoError(t, placeErr.ErrorContext, "PlaceOrder should succeed")
			}
			require.Nil(t, placeErr, "PlaceOrder should succeed")

			t.Log(utility.PrettyPrintJSON(placedOrder))

			require.NotNil(t, placedOrder, "Placed order should not be nil")
			require.NotNil(t, placedOrder.ID, "Placed order should have an ID")
			require.Equal(t, entity3.OrderStatus__WaitingForPayment, placedOrder.Status, "Initial order status should be waiting-for-payment")
			require.Len(t, placedOrder.Items, len(placeOrderPayload.Items), "Number of items should match payload")
			require.NotNil(t, placedOrder.Invoice, "Order should have an associated invoice")
			require.NotZero(t, placedOrder.Invoice.Amount, "Invoice amount should be calculated")
			require.Equal(t, "1", placedOrder.Invoice.ExternalID) // the id of external payment gateway

			// --- Act: Handle Webhook (Failure) ---
			webhookPayloadFail := mock_payment.Invoice{
				ExternalId: strconv.FormatInt(placedOrder.ID, 10),
				Status:     "expired", // Simulate failed/expired payment
				Amount:     float32(placedOrder.Invoice.Amount),
			}
			webhookErrFail := webhookUsecase.HandleWebhook(ctx, webhookPayloadFail)

			// --- Assert: Handle Webhook (Failure) ---
			if webhookErrFail != nil {
				require.NoError(t, webhookErrFail.ErrorContext, "Handling failed webhook should not return an error")
			}
			require.Nil(t, webhookErrFail, "Handling failed webhook should not return an error")

			// --- Act: Get Order (After Failure) ---
			getOrderPayload := entity3.GetOrderDto{
				OrderID: placedOrder.ID,
				UserID:  &userID,
			}
			fetchedOrderFail, getOrderErrFail := getOrderUsecase.GetOrder(ctx, getOrderPayload)

			// --- Assert: Get Order (After Failure) ---
			if getOrderErrFail != nil {
				require.NoError(t, getOrderErrFail.ErrorContext, "GetOrder after failure should succeed")
			}
			require.Nil(t, getOrderErrFail, "GetOrder after failure should succeed")
			require.NotNil(t, fetchedOrderFail, "Fetched order after failure should not be nil")
			require.Equal(t, placedOrder.ID, fetchedOrderFail.ID)
			require.Equal(t, entity3.OrderStatus__Failed, fetchedOrderFail.Status, "Order status should be failed")
			require.NotNil(t, fetchedOrderFail.FailReason, "FailReason should be set for failed order")
			assert.Contains(t, *fetchedOrderFail.FailReason, "expired", "FailReason should indicate expiry/failure") // Check based on webhook status
			require.NotNil(t, fetchedOrderFail.Invoice)
			require.Equal(t, entity.InvoiceStatus__Expired, fetchedOrderFail.Invoice.Status, "Invoice status should be expired/failed")

			for _, item := range fetchedOrderFail.Items {
				require.Equal(t, entity2.SeatStatus__Available, item.TicketSeat.Status)
			}

			availabilities, err = app.availabilityRepository.GetAvailability(ctx, entity2.GetAvailabilityDto{
				TicketSaleID: ticketSale.ID,
			})
			require.NoError(t, err)

			var afterAvailability *entity2.AreaAvailability

			for _, currentAvailability := range availabilities {
				if currentAvailability.TicketAreaID == ticketArea.ID {
					afterAvailability = &currentAvailability
				}
			}

			require.NotNil(t, afterAvailability)
			require.Equal(t, initialAvailability.AvailableSeats, afterAvailability.AvailableSeats)

			// --- Act & Assert: Get Issued Tickets (Should Fail or be Empty) ---
			getTicketsPayload := entity4.GetIssuedTicketDto{
				ID:     placedOrder.ID,
				UserID: &userID,
			}
			issuedTickets, getTicketsErr := getOrderUsecase.GetIssuedTicket(ctx, getTicketsPayload)

			// Check for expected error or empty result depending on implementation
			if getTicketsErr != nil {
				assert.ErrorIs(t, getTicketsErr.ErrorContext, entity4.IssuedTicketNotFoundError, "Expected IssuedTicketNotFoundError for failed order")
				assert.Equal(t, http.StatusNotFound, getTicketsErr.Code)
			} else {
				assert.Empty(t, issuedTickets, "Issued tickets slice should be empty for a failed order")
			}
		})
	}
}
