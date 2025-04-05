package order

import (
	"context"
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"time"
	"tugas-akhir/backend/infrastructure/scylla"
	entity3 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/repository/event"
	"tugas-akhir/backend/internal/idgen"
	"tugas-akhir/backend/internal/orders/entity"
	entity2 "tugas-akhir/backend/internal/payments/entity"
)

type ScyllaOrderRepository struct {
	scylla          *scylla.Scylla
	eventRepository event.EventRepository
	idgen           *idgen.Idgen
}

func NewScyllaOrderRepository(
	scylla *scylla.Scylla,
	eventRepository event.EventRepository,
	idgen *idgen.Idgen,
) *ScyllaOrderRepository {
	return &ScyllaOrderRepository{
		scylla:          scylla,
		idgen:           idgen,
		eventRepository: eventRepository,
	}
}

func (r *ScyllaOrderRepository) PlaceOrder(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, error) {
	if payload.UserID == nil {
		return nil, errors.Wrap(entity.OrderPlacementInternalError, "user id is nil")
	}

	if payload.FirstTicketAreaID == nil {
		return nil, errors.Wrap(entity.OrderPlacementInternalError, "first ticket area id is nil")
	}

	// Generate ID using Sonyflake instead of TimeUUID
	orderID, err := r.idgen.Next()

	if err != nil {
		return nil, errors.Wrap(err, "failed to generate order ID")
	}

	now := time.Now()

	// Create order items UDT list
	var orderItems []entity.OrderItem
	var orderItemsUDT []map[string]interface{}

	for _, item := range payload.Items {
		if item.Price == nil {
			return nil, errors.Wrap(entity.OrderPlacementInternalError, "order item price is nil")
		}

		if item.TicketCategoryID == nil {
			return nil, errors.Wrap(entity.OrderPlacementInternalError, "order item ticket category is nil")
		}

		itemID, err := r.idgen.Next()

		if err != nil {
			return nil, errors.Wrap(err, "failed to generate item ID")
		}

		// Create order item
		orderItem := entity.OrderItem{
			ID:               itemID,
			CustomerName:     item.CustomerName,
			CustomerEmail:    item.CustomerEmail,
			Price:            int64(*item.Price),
			OrderID:          orderID,
			TicketCategoryID: *item.TicketCategoryID,
			TicketSeatID:     *item.TicketSeatID,
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		orderItems = append(orderItems, orderItem)

		// Create UDT map for ScyllaDB
		itemUDT := map[string]interface{}{
			"id":                 itemID,
			"customer_name":      item.CustomerName,
			"customer_email":     item.CustomerEmail,
			"price":              *item.Price,
			"ticket_category_id": *item.TicketCategoryID,
			"ticket_seat_id":     item.TicketSeatID,
			"created_at":         now,
			"updated_at":         now,
		}

		orderItemsUDT = append(orderItemsUDT, itemUDT)
	}

	// Insert into orders table with items as UDT list
	if err := r.scylla.Session.Query(`
		INSERT INTO ticket_system.orders (
			id, status, event_id, ticket_sale_id, first_ticket_area_id, 
			external_user_id, created_at, updated_at, items
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, string(entity.OrderStatus__WaitingForPayment), payload.EventID, payload.TicketSaleID,
		*payload.FirstTicketAreaID, *payload.UserID, now, now, orderItemsUDT,
	).WithContext(ctx).Exec(); err != nil {
		return nil, errors.Wrap(err, "failed to insert order with items")
	}

	// Create order object to return
	order := entity.Order{
		ID:                orderID,
		Status:            entity.OrderStatus__WaitingForPayment,
		EventID:           payload.EventID,
		TicketSaleID:      payload.TicketSaleID,
		FirstTicketAreaID: *payload.FirstTicketAreaID,
		ExternalUserID:    *payload.UserID,
		CreatedAt:         now,
		UpdatedAt:         now,
		Items:             orderItems,
	}

	return &order, nil
}

func (r *ScyllaOrderRepository) GetOrder(ctx context.Context, payload entity.GetOrderDto) (*entity.Order, error) {
	if payload.UserID == nil && !payload.BypassUserID {
		return nil, errors.Wrap(entity.OrderFetchInternalError, "user id is null with no bypass")
	}

	// Get order data including nested items
	var order entity.Order

	// Basic order fields
	var status entity.OrderStatus
	var failReason *string
	var eventID, ticketSaleID, firstTicketAreaID int64
	var externalUserID string
	var createdAt, updatedAt time.Time

	// Invoice fields (now part of the orders table)
	var invoiceID *int64
	var invoiceStatus entity2.InvoiceStatus
	var invoiceExternalID *string
	var invoiceAmount *int
	var invoiceCreatedAt, invoiceUpdatedAt *time.Time

	// Items will be retrieved separately due to ScyllaDB limitations with UDT collections in Go
	if err := r.scylla.Session.Query(`
		SELECT id, status, fail_reason, event_id, ticket_sale_id, 
		       first_ticket_area_id, external_user_id, created_at, updated_at,
		       invoice_id, invoice_status, invoice_amount, invoice_external_id,
		       invoice_created_at, invoice_updated_at
		FROM ticket_system.orders 
		WHERE id = ?`,
		payload.OrderID,
	).WithContext(ctx).Scan(
		&order.ID, &status, failReason, &eventID, &ticketSaleID,
		&firstTicketAreaID, &externalUserID, &createdAt, &updatedAt,
		invoiceID, &invoiceStatus, invoiceAmount, invoiceExternalID,
		invoiceCreatedAt, invoiceUpdatedAt,
	); err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, entity.OrderNotFoundError
		}

		if payload.UserID != nil && !payload.BypassUserID {
			if *payload.UserID != externalUserID {
				return nil, entity.OrderNotFoundError
			}
		}

		return nil, errors.Wrap(err, "failed to get order")
	}

	order.Status = status
	order.FailReason = failReason
	order.EventID = eventID
	order.TicketSaleID = ticketSaleID
	order.FirstTicketAreaID = firstTicketAreaID
	order.ExternalUserID = externalUserID
	order.CreatedAt = createdAt
	order.UpdatedAt = updatedAt

	// Process invoice data
	if invoiceID != nil {
		invoice := entity2.Invoice{
			ID:         *invoiceID,
			Status:     invoiceStatus,
			Amount:     int32(*invoiceAmount),
			ExternalID: *invoiceExternalID,
			OrderID:    payload.OrderID,
			CreatedAt:  *invoiceCreatedAt,
			UpdatedAt:  *invoiceUpdatedAt,
		}
		order.Invoice = &invoice
	}

	// Retrieve order items separately
	// ScyllaDB has limitations with UDT collections in Go, so we'll query the UDT list as individual rows
	var orderItems []entity.OrderItem

	// For this implementation, we'll use a custom query to extract items from the UDT list
	// This approach depends on your Scylla version and gocql capabilities
	iter := r.scylla.Session.Query(`
		SELECT
			items[i].id,
			items[i].customer_name,
			items[i].customer_email,
			items[i].price,
			items[i].ticket_category_id,
			items[i].ticket_seat_id,
			items[i].created_at,
			items[i].updated_at
		FROM ticket_system.orders
		WHERE id = ?
		ALLOW FILTERING`,
		payload.OrderID,
	).WithContext(ctx).Iter()

	var item entity.OrderItem

	for iter.Scan(
		&item.ID, &item.CustomerName, &item.CustomerEmail, &item.Price,
		&item.TicketCategoryID, &item.TicketSeatID,
		&item.CreatedAt, &item.UpdatedAt,
	) {
		item.OrderID = payload.OrderID

		orderItems = append(orderItems, item)

		// Reset for next scan
		item = entity.OrderItem{}
	}

	if err := iter.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to get order items")
	}

	order.Items = orderItems

	// enrich rest of the data
	eventEntity, err := r.eventRepository.GetEvent(ctx, entity3.GetEventDto{
		ID: order.EventID,
	})

	if err != nil {
		return nil, err
	}

	var ticketSale *entity3.TicketSale

	for _, sale := range eventEntity.TicketSales {
		if sale.ID == order.TicketSaleID {
			ticketSale = &sale
		}
	}

	if ticketSale == nil {
		return nil, errors.Wrap(entity.OrderFetchInternalError, "ticket sale not found")
	}

	order.Event = &entity3.Event{
		ID:          eventEntity.ID,
		Name:        eventEntity.Name,
		Location:    eventEntity.Location,
		Description: eventEntity.Description,
		CreatedAt:   eventEntity.CreatedAt,
		UpdatedAt:   eventEntity.UpdatedAt,
	}

	order.TicketSale = &entity3.TicketSale{
		ID:          ticketSale.ID,
		Name:        ticketSale.Name,
		SaleEndAt:   ticketSale.SaleEndAt,
		SaleBeginAt: ticketSale.SaleBeginAt,
		EventID:     ticketSale.EventID,
		CreatedAt:   ticketSale.CreatedAt,
		UpdatedAt:   ticketSale.UpdatedAt,
	}

	// enrich ticket seats data
	for i := 0; i < len(order.Items); i++ {
		item := order.Items[i]

		var id int64
		var seatNumber string
		var status entity3.SeatStatus
		var ticketAreaID int64
		var createdAt, updatedAt time.Time

		found := false

		numberedErr := r.scylla.Session.Query(`
		SELECT id, seat_number, status, ticket_area_id, created_at, updated_at
		FROM ticket_system.ticket_seats_numbered 
		WHERE id = ?`,
			item.TicketSeatID,
		).WithContext(ctx).Scan(
			&id, &seatNumber, &status, &ticketAreaID, &createdAt, &updatedAt,
		)

		if numberedErr != nil && !errors.Is(err, gocql.ErrNotFound) {
			return nil, errors.Wrap(err, "failed to get order item seat")
		} else if numberedErr == nil {
			found = true
		}

		areaErr := r.scylla.Session.Query(`
		SELECT id, seat_number, status, ticket_area_id, created_at, updated_at
		FROM ticket_system.ticket_seats_area 
		WHERE id = ?`,
			item.TicketSeatID,
		).WithContext(ctx).Scan(
			&id, &seatNumber, &status, &ticketAreaID, &createdAt, &updatedAt,
		)

		if areaErr != nil && !errors.Is(err, gocql.ErrNotFound) {
			return nil, errors.Wrap(err, "failed to get order item seat")
		} else if areaErr == nil {
			found = true
		}

		if !found {
			return nil, errors.Wrap(err, "failed to get order item seat. not found")
		}

		var ticketCategory *entity3.TicketCategory
		var ticketArea *entity3.TicketArea

		for _, ticketPackage := range ticketSale.TicketPackages {
			if ticketPackage.TicketCategoryID == item.TicketCategoryID {
				ticketCategory = &ticketPackage.TicketCategory

				for _, area := range ticketPackage.TicketAreas {
					if area.ID == ticketAreaID {
						ticketArea = &area
						break
					}
				}

				break
			}
		}

		if ticketCategory == nil {
			return nil, errors.Wrap(err, "failed to get ticket category for order item")
		}

		if ticketArea == nil {
			return nil, errors.Wrap(err, "failed to get ticket category for ticket area")
		}

		order.Items[i].TicketSeat = &entity3.TicketSeat{
			ID:           id,
			SeatNumber:   seatNumber,
			Status:       status,
			TicketAreaID: ticketAreaID,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			TicketArea: &entity3.TicketArea{
				ID:              ticketArea.ID,
				Type:            ticketArea.Type,
				TicketPackageID: ticketArea.TicketPackageID,
				CreatedAt:       ticketArea.CreatedAt,
				UpdatedAt:       ticketArea.UpdatedAt,
			},
		}

		order.Items[i].TicketCategory = &entity3.TicketCategory{
			ID:        ticketCategory.ID,
			EventID:   ticketCategory.EventID,
			Name:      ticketCategory.Name,
			CreatedAt: ticketCategory.CreatedAt,
			UpdatedAt: ticketCategory.UpdatedAt,
		}
	}

	return &order, nil
}

func (r *ScyllaOrderRepository) UpdateOrderStatus(ctx context.Context, payload entity.UpdateOrderStatusDto) error {
	// Update order status
	if err := r.scylla.Session.Query(`
		UPDATE ticket_system.orders 
		SET status = ?, fail_reason = ?, updated_at = ? 
		WHERE id = ?`,
		payload.Status, payload.FailReason, time.Now(), payload.OrderID,
	).WithContext(ctx).Exec(); err != nil {
		return errors.Wrap(err, "failed to update order status")
	}

	return nil
}
