package booking

import (
	"context"
	"errors"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgconn"
	errors2 "github.com/pkg/errors"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/bookings/entity"
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
	finalSeats := make([]entity2.TicketSeat, 0)
	combinedIDs := make([]int64, 0)

	if len(payload.SeatIDs)+len(payload.TicketAreaIDs) == 0 {
		return nil, errors2.WithStack(errors2.WithMessage(entity.InternalTicketLockError, "total seat in payload must not be zero"))
	}

	if len(payload.SeatIDs) != 0 && len(payload.TicketAreaIDs) != 0 {
		return nil, errors2.WithStack(errors2.WithMessage(entity.InternalTicketLockError, "only either ticket seat or ticket area are allowed at the same time"))
	}

	if len(payload.SeatIDs) > 0 {
		numberedQuery := `
	SELECT id, seat_number, status, ticket_area_id, created_at, updated_at
	FROM ticket_seats
	WHERE id = ANY($1) and status = 'available'
	FOR UPDATE NOWAIT
    `

		numberedSeats := make([]entity2.TicketSeat, 0)

		err := pgxscan.Select(ctx, r.db.GetExecutor(ctx), &numberedSeats, numberedQuery, payload.SeatIDs)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				// PostgreSQL error codes for lock-related issues
				// 55P03 is the error code for "no wait" lock failure
				if pgErr.Code == "55P03" {
					return nil, entity.LockNotAcquiredError
				}
			}
			return nil, err
		}

		if len(numberedSeats) != len(payload.SeatIDs) {
			return nil, errors2.WithStack(errors2.WithMessage(entity.InternalTicketLockError, "the result data length does not match with the param length"))
		}

		finalSeats = append(finalSeats, numberedSeats...)
		combinedIDs = append(combinedIDs, payload.SeatIDs...)
	}

	if len(payload.TicketAreaIDs) > 0 {
		freeSeatedQuery := `
	SELECT id, seat_number, status, ticket_area_id, created_at, updated_at
	FROM ticket_seats
	WHERE ticket_area_id = $1 and status = 'available'
	LIMIT $2
	FOR UPDATE SKIP LOCKED
    `

		areaCountMap := make(map[int64]int)

		for _, area := range payload.TicketAreaIDs {
			_, ok := areaCountMap[area]

			if ok {
				areaCountMap[area]++
			} else {
				areaCountMap[area] = 1
			}
		}

		for key, val := range areaCountMap {
			freeSeatedSeats := make([]entity2.TicketSeat, 0)

			err := pgxscan.Select(ctx, r.db.GetExecutor(ctx), &freeSeatedSeats, freeSeatedQuery, key, val)

			if err != nil {
				return nil, err
			}

			if len(freeSeatedSeats) != val {
				return nil, errors2.WithStack(errors2.WithMessage(entity.LockNotAcquiredError, "cannot acquire locks for the given seats"))
			}

			finalSeats = append(finalSeats, freeSeatedSeats...)

			for _, seat := range freeSeatedSeats {
				combinedIDs = append(combinedIDs, seat.ID)
			}
		}
	}

	// update status to on hold
	updateQuery := `
	UPDATE ticket_seats
	SET status = 'on-hold'
	WHERE id = ANY($1) and status = 'available'
    `

	tag, err := r.db.GetExecutor(ctx).Exec(ctx, updateQuery, combinedIDs)

	if err != nil {
		return nil, err
	}

	if tag.RowsAffected() != int64(len(combinedIDs)) {
		return nil, errors2.WithStack(errors2.WithMessage(entity.InternalTicketLockError, "the updated data length does not match with the param length"))
	}

	return finalSeats, nil
}

func (r *PGBookingRepository) UpdateSeatStatus(ctx context.Context, payload entity.UpdateSeatStatusDto) error {
	query := `
	UPDATE ticket_seats
	SET status = $1, updated_at = now()
	WHERE id = ANY($2)
    `

	_, err := r.db.GetExecutor(ctx).Exec(ctx, query, payload.Status, payload.SeatIDs)

	return err
}
