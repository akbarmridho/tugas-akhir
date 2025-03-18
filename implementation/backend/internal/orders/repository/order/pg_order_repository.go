package order

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
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

	querier := r.db.GetExecutor(ctx)

	orderQuery := `
	INSERT INTO orders(external_user_id, first_ticket_area_id, status)
	VALUES ($1, $2, 'waiting-for-payment')
	RETURNING id, status, fail_reason, first_ticket_area_id, external_user_id, created_at, updated_at
    `

	var order entity.Order

	err := pgxscan.Get(ctx, querier, &order, orderQuery, *payload.UserID, *payload.FirstTicketAreaID)

	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "no rows returned for order"))
		}

		return nil, err
	}

	orderItems := make([]entity.OrderItem, 0)

	orderItemQuery := `
	INSERT INTO order_items(customer_name, customer_email, price, order_id, ticket_seat_id) VALUES
    `

	args := []interface{}{}

	for i, item := range payload.Items {
		if i > 0 && i != (len(payload.Items)-1) {
			orderItemQuery += ", "
		}

		if item.Price == nil {
			return nil, errors.WithStack(errors.WithMessage(entity.OrderPlacementInternalError, "order item price is nil"))
		}

		paramOffset := i * 5
		orderItemQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", paramOffset+1, paramOffset+2, paramOffset+3, paramOffset+4, paramOffset+5)
		args = append(args, item.CustomerName, item.CustomerEmail, *item.Price, order.ID, item.TicketSeatID)
	}

	orderItemQuery += " RETURNING id, customer_name, customer_email, price, order_id, ticket_seat_id, created_at, updated_at"

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

// todo get order

// todo update order (for webhooks handle)
