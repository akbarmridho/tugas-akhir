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

	if payload.TicketAreaID == nil {
		return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "first ticket area id is nil"))
	}
	querier := r.db.GetExecutor(ctx)

	orderQuery := `
	INSERT INTO orders(external_user_id, ticket_area_id, status, ticket_sale_id, event_id)
	VALUES ($1, $2, 'waiting-for-payment', $3, $4)
	RETURNING *
    `

	var order entity.Order

	err := pgxscan.Get(ctx, querier, &order, orderQuery, *payload.UserID, *payload.TicketAreaID, payload.TicketSaleID, payload.EventID)

	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "no rows returned for order"))
		}

		return nil, err
	}

	orderItems := make([]entity.OrderItem, 0)

	orderItemQuery := `
	INSERT INTO order_items(customer_name, customer_email, price, order_id, ticket_seat_id, ticket_category_id, ticket_area_id) VALUES
    `

	args := []interface{}{}

	for i, item := range payload.Items {
		if i > 0 {
			orderItemQuery += ", "
		}

		if item.Price == nil {
			return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "order item price is nil"))
		}

		if item.TicketCategoryID == nil {
			return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "order item ticket category is nil"))
		}

		paramOffset := i * 7
		orderItemQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			paramOffset+1, paramOffset+2, paramOffset+3, paramOffset+4, paramOffset+5, paramOffset+6, paramOffset+7)
		args = append(args, item.CustomerName, item.CustomerEmail, *item.Price, order.ID, item.TicketSeatID, *item.TicketCategoryID, order.TicketAreaID)
	}

	orderItemQuery += " RETURNING id, customer_name, customer_email, price, order_id, ticket_seat_id, ticket_category_id, created_at, updated_at"

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
	if payload.UserID == nil && !payload.BypassUserID {
		return nil, entity.OrderFetchInternalError
	}

	var order entity.Order

	query := `
		WITH order_base AS (
			SELECT
				o.id, o.status, o.fail_reason, o.event_id, o.ticket_sale_id,
				o.ticket_area_id, o.external_user_id, o.created_at, o.updated_at
			FROM orders o
			WHERE o.id = $1 AND (o.external_user_id = $2 OR $3::boolean)
		),
		invoice_details AS (
			SELECT
				i.id, i.status, i.amount, i.external_id, i.order_id, i.ticket_area_id AS invoice_ticket_area_id,
				i.created_at, i.updated_at
			FROM invoices i
			JOIN order_base ob ON i.order_id = ob.id AND i.ticket_area_id = ob.ticket_area_id
		),
		event_details AS (
			SELECT
				e.id, e.name, e.location, e.description, e.created_at, e.updated_at
			FROM events e
			JOIN order_base ob ON e.id = ob.event_id
		),
		ticket_sale_details AS (
			SELECT
				ts.id, ts.name, ts.sale_begin_at, ts.sale_end_at, ts.event_id,
				ts.created_at, ts.updated_at
			FROM ticket_sales ts
			JOIN order_base ob ON ts.id = ob.ticket_sale_id
		),
		order_items_enriched AS (
			SELECT
				oi.id AS item_id, oi.customer_name, oi.customer_email, oi.price, oi.order_id AS item_order_id,
				oi.ticket_category_id, oi.ticket_seat_id AS item_ticket_seat_id,
				oi.ticket_area_id AS item_ticket_area_id,
				oi.created_at AS item_created_at, oi.updated_at AS item_updated_at,
				
				tc.id AS category_id, tc.name AS category_name, tc.event_id AS category_event_id,
				tc.created_at AS category_created_at, tc.updated_at AS category_updated_at,
				
				ts_seat.id AS seat_id, ts_seat.seat_number, ts_seat.status AS seat_status,
				ts_seat.ticket_area_id AS seat_ticket_area_id,
				ts_seat.created_at AS seat_created_at, ts_seat.updated_at AS seat_updated_at,
				
				ta.id AS area_id, ta.type AS area_type, ta.ticket_package_id AS area_ticket_package_id,
				ta.created_at AS area_created_at, ta.updated_at AS area_updated_at
				
			FROM order_items oi
			JOIN order_base ob ON oi.order_id = ob.id AND oi.ticket_area_id = ob.ticket_area_id
			LEFT JOIN ticket_categories tc ON oi.ticket_category_id = tc.id
			LEFT JOIN ticket_seats ts_seat ON oi.ticket_seat_id = ts_seat.id AND oi.ticket_area_id = ts_seat.ticket_area_id
			LEFT JOIN ticket_areas ta ON ts_seat.ticket_area_id = ta.id
			WHERE ob.id IS NOT NULL 
		),
		aggregated_order_items AS (
			SELECT
				item_order_id,
				jsonb_agg(
					jsonb_build_object(
						'id', item_id,
						'customerName', customer_name,
						'customerEmail', customer_email,
						'price', price,
						'orderId', item_order_id,
						'ticketCategoryId', ticket_category_id,
						'ticketSeatId', seat_id,
						'ticketAreaId', item_ticket_area_id,
						'createdAt', item_created_at,
						'updatedAt', item_updated_at,
						'ticketSeat', CASE WHEN seat_id IS NOT NULL THEN jsonb_build_object(
							'id', seat_id,
							'seatNumber', seat_number,
							'status', seat_status,
							'ticketAreaId', seat_ticket_area_id,
							'createdAt', seat_created_at,
							'updatedAt', seat_updated_at,
							'ticketArea', CASE WHEN area_id IS NOT NULL THEN jsonb_build_object(
								'id', area_id,
								'type', area_type,
								'ticketPackageId', area_ticket_package_id,
								'createdAt', area_created_at,
								'updatedAt', area_updated_at
							) ELSE NULL END
						) ELSE NULL END,
						'ticketCategory', CASE WHEN category_id IS NOT NULL THEN jsonb_build_object(
							'id', category_id,
							'name', category_name,
							'eventId', category_event_id,
							'createdAt', category_created_at,
							'updatedAt', category_updated_at
						) ELSE NULL END
					)
				ORDER BY item_id
				) AS items_json
			FROM order_items_enriched
			GROUP BY item_order_id
		)
		SELECT
			jsonb_build_object(
				'id', o.id,
				'status', o.status,
				'failReason', o.fail_reason,
				'eventId', o.event_id,
				'ticketSaleId', o.ticket_sale_id,
				'ticketAreaId', o.ticket_area_id,
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
						'ticketAreaId', i.invoice_ticket_area_id,
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
				
				'ticketSale', CASE WHEN tsale.id IS NOT NULL THEN
					jsonb_build_object(
						'id', tsale.id,
						'name', tsale.name,
						'saleBeginAt', tsale.sale_begin_at,
						'saleEndAt', tsale.sale_end_at,
						'eventId', tsale.event_id,
						'createdAt', tsale.created_at,
						'updatedAt', tsale.updated_at
					)
				ELSE NULL END,
				
				'items', COALESCE(aoi.items_json, '[]'::jsonb)
			) as order_json
		FROM order_base o
		LEFT JOIN invoice_details i ON o.id = i.order_id AND o.ticket_area_id = i.invoice_ticket_area_id
		LEFT JOIN event_details e ON o.event_id = e.id
		LEFT JOIN ticket_sale_details tsale ON o.ticket_sale_id = tsale.id
		LEFT JOIN aggregated_order_items aoi ON o.id = aoi.item_order_id
		WHERE o.id IS NOT NULL;
    `

	var orderJSON json.RawMessage
	err := r.db.GetExecutor(ctx).QueryRow(ctx, query, payload.OrderID, payload.UserID, payload.BypassUserID).Scan(&orderJSON)
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

func (r *PGOrderRepository) UpdateOrderStatus(ctx context.Context, payload entity.UpdateOrderStatusDto) error {
	query := `
	UPDATE orders
	SET status = $1, fail_reason = $2, updated_at = now()
	WHERE id = $3
    `

	rows, err := r.db.GetExecutor(ctx).Exec(ctx, query, payload.Status, payload.FailReason, payload.OrderID)

	if rows.RowsAffected() == 0 {
		return errors.WithStack(entity.OrderNotFoundError)
	}

	return err
}
