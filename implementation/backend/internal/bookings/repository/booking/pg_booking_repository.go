package booking

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/bookings/entity"
	entity2 "tugas-akhir/backend/internal/events/entity"
)

type PGBookingInterface struct {
	db *postgres.Postgres
}

func NewPGBookingInterface(db *postgres.Postgres) *PGBookingInterface {
	return &PGBookingInterface{
		db: db,
	}
}

func (r *PGBookingInterface) Book(ctx context.Context, payload entity.BookingRequestDto) ([]entity2.TicketSeat, error) {
	query := `
	SELECT id, seat_number, status, ticket_area_id, created_at, updated_at
	FROM ticket_seats
	WHERE id = ANY($1) and status = 'available'
	FOR UPDATE NOWAIT
    `

	seats := make([]entity2.TicketSeat, 0)

	err := pgxscan.Select(ctx, r.db.GetExecutor(ctx), &seats, query, payload.SeatIDs)

	if err != nil {
		return nil, err
	}

	if len(seats) != len(payload.SeatIDs) {
		return nil, entity.ResultDataLengthNotMatch
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
		return nil, entity.UpdatedDataLengthNotMatch
	}

	return seats, nil
}
