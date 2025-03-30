package booking

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/bookings/entity"
	"tugas-akhir/backend/internal/bookings/service"
	entity2 "tugas-akhir/backend/internal/events/entity"
)

type PGBookingRepository struct {
	db *postgres.Postgres
}

func NewPGBookingRepository(db *postgres.Postgres) *PGBookingRepository {
	return &PGBookingRepository{
		db: db,
	}
}

func (r *PGBookingRepository) Book(ctx context.Context, payload entity.BookingRequestDto) ([]entity2.TicketSeat, error) {
	query := `
	SELECT id, seat_number, status, ticket_area_id, created_at, updated_at
	FROM ticket_seats
	WHERE id = ANY($1) and status = 'available'
	FOR UPDATE NOWAIT
    `

	seats := make([]entity2.TicketSeat, 0)

	err := pgxscan.Select(ctx, r.db.GetExecutor(ctx), &seats, query, payload.SeatIDs)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			// PostgreSQL error codes for lock-related issues
			// 55P03 is the error code for "no wait" lock failure
			if pgErr.Code == "55P03" {
				return nil, entity.LockNotAcquiredError
			}
		}
		return nil, err
	}

	if len(seats) != len(payload.SeatIDs) {
		return nil, errors.WithStack(errors.WithMessage(entity.InternalTicketLockError, "the result data length does not match with the param length"))
	}

	// update status to on hold
	updateQuery := `
	UPDATE ticket_seats
	SET status = 'on-hold'
	WHERE id = ANY($1) and status = 'available'
    `

	tag, err := r.db.GetExecutor(ctx).Exec(ctx, updateQuery, payload.SeatIDs)

	if err != nil {
		return nil, err
	}

	if tag.RowsAffected() != int64(len(payload.SeatIDs)) {
		return nil, errors.WithStack(errors.WithMessage(entity.InternalTicketLockError, "the updated data length does not match with the param length"))
	}

	return seats, nil
}

func (r *PGBookingRepository) PublishIssuedTickets(ctx context.Context, payload entity.PublishIssuedTicketDto) error {
	query := `
	INSERT INTO issued_tickets(serial_number, holder_name, seat_id, order_id, order_item_id, name, description) VALUES
    `

	if len(payload.Items) != len(payload.SeatInfos) {
		return errors.WithMessage(entity.IssueTicketError, "payload items and seat info length is different")
	}

	args := []interface{}{}

	for i, item := range payload.Items {
		if i > 0 && i != (len(payload.Items)-1) {
			query += ", "
		}

		info := payload.SeatInfos[i]

		serialNumber, err := service.GenerateSerialNumber(item)

		if err != nil {
			return err
		}

		issuedTicketDescription := ""

		if info.SeatType == entity2.AreaType__FreeStanding {
			issuedTicketDescription = fmt.Sprintf("%s (Free Standing)", info.CategoryName)
		} else if info.SeatType == entity2.AreaType__NumberedSeating {
			issuedTicketDescription = fmt.Sprintf("%s - Number %s", info.CategoryName, info.SeatNumber)
		}

		paramOffset := i * 7
		query += fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			paramOffset+1,
			paramOffset+2,
			paramOffset+3,
			paramOffset+4,
			paramOffset+5,
			paramOffset+6,
			paramOffset+7,
		)

		args = append(args,
			serialNumber,
			item.CustomerName,
			item.TicketSeatID,
			item.OrderID,
			item.ID,
			fmt.Sprintf("%s - %s", payload.EventName, payload.TicketSaleName),
			issuedTicketDescription,
		)
	}

	_, err := r.db.GetExecutor(ctx).Exec(ctx, query, args...)

	return err
}

func (r *PGBookingRepository) GetIssuedTickets(ctx context.Context, payload entity.GetIssuedTicketDto) ([]entity.IssuedTicket, error) {
	var count int

	countQuery := `
        SELECT COUNT(*) 
        FROM orders 
        WHERE id = $1 AND external_user_id = $2
    `

	err := r.db.GetExecutor(ctx).QueryRow(ctx, countQuery, payload.ID, payload.UserID).Scan(&count)

	if err != nil {
		return nil, errors.WithStack(errors.WithMessage(entity.IssuedTicketFetchError, "cannot get the order count"))
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
			it.seat_id, 
			it.order_id, 
			it.order_item_id, 
			it.created_at, 
			it.updated_at,
			ts.id AS "ticketSeat.id",
			ts.seat_number AS "ticketSeat.seatNumber",
			ts.status AS "ticketSeat.status",
			ts.ticket_area_id AS "ticketSeat.ticketAreaId",
			ts.created_at AS "ticketSeat.createdAt",
			ts.updated_at AS "ticketSeat.updatedAt"
		FROM issued_tickets it
		JOIN ticket_seats ts ON it.seat_id = ts.id
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
