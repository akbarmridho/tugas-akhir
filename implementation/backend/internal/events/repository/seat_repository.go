package repository

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/events/entity"
)

type SeatRepository struct {
	db *postgres.Postgres
}

func NewSeatRepository(
	db *postgres.Postgres,
) *SeatRepository {
	return &SeatRepository{
		db: db,
	}
}

func (r *EventRepository) GetSeats(ctx context.Context, payload entity.GetSeatsDto) ([]entity.TicketSeat, error) {
	query := `
	SELECT *
	FROM ticket_seats
	WHERE ticket_area_id = $1
	ORDER BY seat_number
	`

	result := make([]entity.TicketSeat, 0)

	err := pgxscan.Select(
		ctx,
		r.db.GetExecutor(ctx),
		&result,
		query,
		payload.TicketAreaID,
	)

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, entity.SeatNotFoundError
	}

	return result, nil
}
