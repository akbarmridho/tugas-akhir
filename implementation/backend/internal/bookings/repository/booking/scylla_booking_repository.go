package booking

import (
	"context"
	"errors"
	"github.com/gocql/gocql"
	errors2 "github.com/pkg/errors"
	"time"
	"tugas-akhir/backend/infrastructure/scylla"
	"tugas-akhir/backend/internal/bookings/entity"
	entity2 "tugas-akhir/backend/internal/events/entity"
)

type ScyllaBookingRepository struct {
	scylla *scylla.Scylla
}

func NewScyllaBookingRepository(scylla *scylla.Scylla) *ScyllaBookingRepository {
	return &ScyllaBookingRepository{
		scylla: scylla,
	}
}

func (r *ScyllaBookingRepository) Book(ctx context.Context, payload entity.BookingRequestDto) ([]entity2.TicketSeat, error) {
	finalSeats := make([]entity2.TicketSeat, 0)

	// Process numbered (specific) seats
	if len(payload.SeatIDs) > 0 {
		batch := r.scylla.Session.Batch(gocql.LoggedBatch)

		// First, get all seats and check if they're available
		for _, seatID := range payload.SeatIDs {
			var seatStatus string
			var seatNumber string
			var ticketAreaID int64
			var createdAt, updatedAt time.Time

			// Use the new numbered seats table
			err := r.scylla.Session.Query(`
				SELECT seat_number, status, ticket_area_id, created_at, updated_at
				FROM ticket_system.ticket_seats_numbered
				WHERE id = ?`,
				seatID,
			).WithContext(ctx).Scan(&seatNumber, &seatStatus, &ticketAreaID, &createdAt, &updatedAt)

			if err != nil {
				if errors.Is(err, gocql.ErrNotFound) {
					return nil, errors2.Wrapf(entity.InternalTicketLockError, "seat %d not found", seatID)
				}
				return nil, errors2.Wrap(err, "failed to query seat")
			}

			if seatStatus != string(entity2.SeatStatus__Available) {
				return nil, errors2.Wrapf(entity.InternalTicketLockError, "seat %d is not available", seatID)
			}

			// Add to CAS batch for atomic update on the numbered seats table
			batch.Query(`
				UPDATE ticket_system.ticket_seats_numbered
				SET status = ?, updated_at = ?
				WHERE id = ?
				IF status = 'available'`,
				string(entity2.SeatStatus__OnHold), time.Now(), seatID,
			)

			// Add to final seats list
			seat := entity2.TicketSeat{
				ID:           seatID,
				SeatNumber:   seatNumber,
				Status:       entity2.SeatStatus__OnHold,
				TicketAreaID: ticketAreaID,
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
			}

			finalSeats = append(finalSeats, seat)
		}

		// Execute the batch with LWT
		applied, _, err := r.scylla.Session.ExecuteBatchCAS(batch)

		if err != nil {
			return nil, errors2.Wrap(err, "failed to update seats")
		}

		if !applied {
			return nil, entity.LockNotAcquiredError
		}
	}

	// Process free-standing areas (non-numbered seats)
	if len(payload.TicketAreaIDs) > 0 {
		areaCountMap := make(map[int64]int)

		for _, area := range payload.TicketAreaIDs {
			areaCountMap[area]++
		}

		for areaID, count := range areaCountMap {
			// Get available seats from this area using the new table
			iter := r.scylla.Session.Query(`
				SELECT id, seat_number, status, created_at, updated_at
				FROM ticket_system.ticket_seats_area
				WHERE ticket_area_id = ? AND status = 'available'
				LIMIT ?`,
				areaID, count,
			).WithContext(ctx).Iter()

			seats := make([]entity2.TicketSeat, 0, count)
			seatIDs := make([]int64, 0, count)

			var id int64
			var seatNumber string
			var status string
			var createdAt, updatedAt time.Time

			for iter.Scan(&id, &seatNumber, &status, &createdAt, &updatedAt) {
				seats = append(seats, entity2.TicketSeat{
					ID:           id,
					SeatNumber:   seatNumber,
					Status:       status,
					TicketAreaID: areaID, // We already know the ticket area ID
					CreatedAt:    createdAt,
					UpdatedAt:    updatedAt,
				})

				seatIDs = append(seatIDs, id)
			}

			if err := iter.Close(); err != nil {
				return nil, errors.Wrap(err, "failed to iterate available seats")
			}

			if len(seats) < count {
				return nil, errors.Wrapf(entity.InternalTicketLockError,
					"not enough available seats in area %d: requested %d, found %d",
					areaID, count, len(seats))
			}

			// Update all the seats to on-hold
			batch := r.scylla.Session.NewBatch(gocql.LoggedBatch)
			now := time.Now()

			for _, seatID := range seatIDs {
				// Update the seat status in ticket_seats_area table
				batch.Query(`
					UPDATE ticket_system.ticket_seats_area
					SET status = 'on-hold', updated_at = ?
					WHERE ticket_area_id = ? AND id = ?
					IF status = 'available'`,
					now, areaID, seatID,
				)
			}

			// Execute batch with CAS
			applied, _, err := r.scylla.Session.ExecuteBatchCAS(batch)
			if err != nil {
				return nil, errors.Wrap(err, "failed to update seats batch")
			}

			if !applied {
				return nil, entity.LockNotAcquiredError
			}

			// Update status in return objects and add to finalSeats
			for i := range seats {
				seats[i].Status = "on-hold"
			}

			finalSeats = append(finalSeats, seats...)
		}
	}

	return finalSeats, nil
}
