package order

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/orders/entity"
)

type PGOrderRepository struct {
	db *postgres.Postgres
}

func NewPGOrderRepository(db *postgres.Postgres) *PGOrderRepository {
	return &PGOrderRepository{
		db: db,
	}
}

func (r *PGOrderRepository) PlaceOrder(ctx context.Context, payload entity.PlaceOrderDto) (*entity.Order, error) {
	if payload.UserID == nil {
		return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "user id is nil"))
	}

	if payload.FirstTicketAreaID == nil {
		return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "first ticket area id is nil"))
	}

	if payload.EventID == nil {
		return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "event id is nil"))
	}

	if payload.TicketSaleID == nil {
		return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "ticket sale id is nil"))
	}

	querier := r.db.GetExecutor(ctx)

	orderQuery := `
	INSERT INTO orders(external_user_id, first_ticket_area_id, status, ticket_sale_id, event_id)
	VALUES ($1, $2, 'waiting-for-payment', $3, $4)
	RETURNING *
    `

	var order entity.Order

	err := pgxscan.Get(ctx, querier, &order, orderQuery, *payload.UserID, *payload.FirstTicketAreaID, *payload.TicketSaleID, *payload.EventID)

	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "no rows returned for order"))
		}

		return nil, err
	}

	orderItems := make([]entity.OrderItem, 0)

	orderItemQuery := `
	INSERT INTO order_items(customer_name, customer_email, price, order_id, ticket_seat_id, ticket_category_id) VALUES
    `

	args := []interface{}{}

	for i, item := range payload.Items {
		if i > 0 && i != (len(payload.Items)-1) {
			orderItemQuery += ", "
		}

		if item.Price == nil {
			return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "order item price is nil"))
		}

		if item.TicketCategoryID == nil {
			return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "order item ticket category is nil"))
		}

		paramOffset := i * 6
		orderItemQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", paramOffset+1, paramOffset+2, paramOffset+3, paramOffset+4, paramOffset+5, paramOffset+6)
		args = append(args, item.CustomerName, item.CustomerEmail, *item.Price, order.ID, item.TicketSeatID, *item.TicketCategoryID)
	}

	orderItemQuery += " RETURNING *"

	err = pgxscan.Select(ctx, querier, &orderItems, orderItemQuery, args...)

	if err != nil {
		return nil, err
	}

	if len(orderItems) != len(payload.Items) {
		return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "returned order items length is different form payload length"))
	}

	order.Items = orderItems

	return &order, nil
}

func (r *PGOrderRepository) GetOrder(ctx context.Context, payload entity.GetOrderDto) (*entity.Order, error) {
	if payload.UserID == nil {
		return nil, entity.OrderFetchInternalError
	}

	var order entity.Order

	query := `
		WITH order_data AS (
			SELECT
				o.id, o.status, o.fail_reason, o.event_id, o.ticket_sale_id,
				o.first_ticket_area_id, o.external_user_id, o.created_at, o.updated_at
			FROM orders o
			WHERE o.id = $1 AND o.external_user_id = $2
		),
		invoice_data AS (
			SELECT
				i.id, i.status, i.amount, i.external_id, i.order_id,
				i.created_at, i.updated_at
			FROM invoices i
			JOIN order_data od ON i.order_id = od.id
		),
		event_data AS (
			SELECT
				e.id, e.name, e.location, e.description, e.created_at, e.updated_at
			FROM events e
			JOIN order_data od ON e.id = od.event_id
		),
		ticket_sale_data AS (
			SELECT
				ts.id, ts.name, ts.sale_begin_at, ts.sale_end_at, ts.event_id,
				ts.created_at, ts.updated_at
			FROM ticket_sales ts
			JOIN order_data od ON ts.id = od.ticket_sale_id
		),
		order_items_data AS (
			SELECT
				oi.id, oi.customer_name, oi.customer_email, oi.price, oi.order_id,
				oi.ticket_category_id, oi.ticket_seat_id, oi.created_at, oi.updated_at
			FROM order_items oi
			JOIN order_data od ON oi.order_id = od.id
		)
		SELECT
			jsonb_build_object(
				'id', o.id,
				'status', o.status,
				'failReason', o.fail_reason,
				'eventId', o.event_id,
				'ticketSaleId', o.ticket_sale_id,
				'firstTicketAreaId', o.first_ticket_area_id,
				'externalUserId', o.external_user_id,
				'createdAt', o.created_at,
				'updatedAt', o.updated_at,
				
				'invoice', CASE WHEN i.id IS NOT NULL THEN
					jsonb_build_object(
						'id', i.id,
						'status', i.status,
						'amount', i.amount,
						'externalId', i.external_id,
						'orderId', i.order_id,
						'createdAt', i.created_at,
						'updatedAt', i.updated_at
					)
				ELSE NULL END,
				
				'event', CASE WHEN e.id IS NOT NULL THEN
					jsonb_build_object(
						'id', e.id,
						'name', e.name,
						'location', e.location,
						'description', e.description,
						'createdAt', e.created_at,
						'updatedAt', e.updated_at
					)
				ELSE NULL END,
				
				'ticketSale', CASE WHEN ts.id IS NOT NULL THEN
					jsonb_build_object(
						'id', ts.id,
						'name', ts.name,
						'saleBeginAt', ts.sale_begin_at,
						'saleEndAt', ts.sale_end_at,
						'eventId', ts.event_id,
						'createdAt', ts.created_at,
						'updatedAt', ts.updated_at
					)
				ELSE NULL END,
				
				'items', (
					SELECT jsonb_agg(items)
					FROM (
						SELECT
							jsonb_build_object(
								'id', oi.id,
								'customerName', oi.customer_name,
								'customerEmail', oi.customer_email,
								'price', oi.price,
								'orderId', oi.order_id,
								'ticketCategoryId', oi.ticket_category_id,
								'ticketSeatId', oi.ticket_seat_id,
								'createdAt', oi.created_at,
								'updatedAt', oi.updated_at,
								
								'ticketSeat', (
									SELECT jsonb_build_object(
										'id', ts.id,
										'seatNumber', ts.seat_number,
										'status', ts.status,
										'ticketAreaId', ts.ticket_area_id,
										'createdAt', ts.created_at,
										'updatedAt', ts.updated_at,
										'ticketArea', (
											SELECT jsonb_build_object(
												'id', ta.id,
												'type', ta.type,
												'ticketPackageId', ta.ticket_package_id,
												'createdAt', ta.created_at,
												'updatedAt', ta.updated_at
											)
											FROM ticket_areas ta
											WHERE ta.id = ts.ticket_area_id
										)
									)
									FROM ticket_seats ts
									WHERE ts.id = oi.ticket_seat_id
								),
								
								'ticketCategory', (
									SELECT jsonb_build_object(
										'id', tc.id,
										'name', tc.name,
										'eventId', tc.event_id,
										'createdAt', tc.created_at,
										'updatedAt', tc.updated_at
									)
									FROM ticket_categories tc
									WHERE tc.id = oi.ticket_category_id
								)
							) as items
						FROM order_items_data oi
					) subq
				)
			) as order_json
		FROM order_data o
		LEFT JOIN invoice_data i ON true
		LEFT JOIN event_data e ON true
		LEFT JOIN ticket_sale_data ts ON true
    `

	var orderJSON json.RawMessage
	err := r.db.GetExecutor(ctx).QueryRow(ctx, query, payload.OrderID, *payload.UserID).Scan(&orderJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.OrderNotFoundError
		}
		return nil, err
	}

	if err := json.Unmarshal(orderJSON, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// todo update order (for webhooks handle)
