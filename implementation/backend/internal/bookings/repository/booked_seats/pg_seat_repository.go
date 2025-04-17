package booked_seats

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	errors2 "github.com/pkg/errors"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/bookings/service"
	entity2 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/pkg/cursor_iterator"
)

type PGBookedSeatRepository struct {
	db        *postgres.Postgres
	generator *service.SerialNumberGenerator
}

func NewPGBookedSeatRepository(
	db *postgres.Postgres,
	generator *service.SerialNumberGenerator,
) *PGBookedSeatRepository {
	return &PGBookedSeatRepository{
		db:        db,
		generator: generator,
	}
}

func (r *PGBookedSeatRepository) PublishIssuedTickets(ctx context.Context, payload entity.PublishIssuedTicketDto) error {
	query := `
	INSERT INTO issued_tickets(serial_number, holder_name, ticket_seat_id, order_id, order_item_id, name, description, ticket_area_id) VALUES
    `

	if len(payload.Items) != len(payload.SeatInfos) {
		return errors2.WithMessage(entity.IssueTicketError, "payload items and seat info length is different")
	}

	args := []interface{}{}

	for i, item := range payload.Items {
		if i > 0 {
			query += ", "
		}

		info := payload.SeatInfos[i]

		serialNumber, err := r.generator.Generate(item)

		if err != nil {
			return err
		}

		issuedTicketDescription := ""

		if info.SeatType == entity2.AreaType__FreeStanding {
			issuedTicketDescription = fmt.Sprintf("%s (Free Standing)", info.CategoryName)
		} else if info.SeatType == entity2.AreaType__NumberedSeating {
			issuedTicketDescription = fmt.Sprintf("%s - Number %s", info.CategoryName, info.SeatNumber)
		}

		paramOffset := i * 8
		query += fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			paramOffset+1,
			paramOffset+2,
			paramOffset+3,
			paramOffset+4,
			paramOffset+5,
			paramOffset+6,
			paramOffset+7,
			paramOffset+8,
		)

		args = append(args,
			serialNumber,
			item.CustomerName,
			item.TicketSeatID,
			item.OrderID,
			item.ID,
			fmt.Sprintf("%s - %s", payload.EventName, payload.TicketSaleName),
			issuedTicketDescription,
			payload.TicketAreaID,
		)
	}

	_, err := r.db.GetExecutor(ctx).Exec(ctx, query, args...)

	return err
}

func (r *PGBookedSeatRepository) GetIssuedTickets(ctx context.Context, payload entity.GetIssuedTicketDto) ([]entity.IssuedTicket, error) {
	var count int

	countQuery := `
        SELECT COUNT(*) 
        FROM orders 
        WHERE id = $1 AND external_user_id = $2
    `

	err := r.db.GetExecutor(ctx).QueryRow(ctx, countQuery, payload.ID, payload.UserID).Scan(&count)

	if err != nil {
		return nil, errors2.WithStack(errors2.WithMessage(entity.IssuedTicketFetchError, "cannot get the order count"))
	}

	if count == 0 {
		return nil, entity.IssuedTicketNotFoundError
	}

	query := `
		SELECT 
			it.id, 
			it.serial_number, 
			it.holder_name, 
			it.name, 
			it.description, 
			it.ticket_seat_id, 
			it.order_id, 
			it.order_item_id, 
			it.created_at, 
			it.updated_at,
			ts.id AS "ticket_seat.id",
			ts.seat_number AS "ticket_seat.seat_number",
			ts.status AS "ticket_seat.status",
			ts.ticket_area_id AS "ticket_seat.ticket_area_id",
			ts.created_at AS "ticket_seat.created_at",
			ts.updated_at AS "ticket_seat.updated_at"
		FROM issued_tickets it
		JOIN ticket_seats ts ON it.ticket_seat_id = ts.id
		WHERE it.order_id = $1
    `

	result := make([]entity.IssuedTicket, 0)

	err = pgxscan.Select(ctx, r.db.GetExecutor(ctx), &result, query, payload.ID)

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, entity.IssuedTicketNotFoundError
	}

	return result, nil
}

func (r *PGBookedSeatRepository) IterSeats(ctx context.Context) ([]entity2.TicketSeat, *cursor_iterator.CursorIterator, error) {
	query := `
	SELECT 
            ts.id, ts.seat_number, ts.status, ts.ticket_area_id, ts.created_at, ts.updated_at,
            ta.id AS "ticket_area.id", 
            ta.type AS "ticket_area.type", 
            ta.ticket_package_id AS "ticket_area.ticket_package_id", 
            ta.created_at AS "ticket_area.created_at", 
            ta.updated_at AS "ticket_area.updated_at"
        FROM 
            ticket_seats ts
        JOIN 
            ticket_areas ta ON ts.ticket_area_id = ta.id
    `

	result := make([]entity2.TicketSeat, 100)

	iter, err := cursor_iterator.NewCursorIterator(r.db.Pool, result, query)

	if err != nil {
		return nil, nil, err
	}

	return result, iter, err
}
