package sanity

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/events/entity"
)

type PGCheck struct {
	db *postgres.Postgres
}

func (s *PGCheck) GetAvailability(ctx context.Context) (*AvailabilityCheck, error) {
	raw := make([]DBAvailabilityRow, 0)

	err := pgxscan.Select(ctx, s.db.GetExecutor(ctx), &raw, `
		SELECT seat_status, count(*) as total
		FROM ticket_seats
		GROUP BY seat_status
	`)

	if err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return nil, fmt.Errorf("query result must have more than one row")
	}

	result := AvailabilityCheck{
		Count:       0,
		Available:   0,
		Unavailable: 0,
	}

	for _, row := range raw {
		if row.SeatStatus == entity.SeatStatus__Available {
			result.Available += row.Total
		} else if row.SeatStatus == entity.SeatStatus__Sold {
			result.Unavailable += row.Total
		} else if row.SeatStatus == entity.SeatStatus__OnHold {
			result.Unavailable += row.Total
		}

		result.Count += row.Total
	}

	return &result, nil
}

func (s *PGCheck) CheckDoubleOrder(ctx context.Context) (*DoubleOrderCheck, error) {
	// sanity check for double order (one seat sold more than once)
	raw := make([]DoubleOrderCheck, 0)

	err := pgxscan.Select(ctx, s.db.GetExecutor(ctx), &raw, `
		SELECT COUNT(*) as total
		FROM (
			SELECT 
				ticket_area_id,
				ticket_seat_id
			FROM 
				order_items
			INNER JOIN 
				orders ON order_items.ticket_area_id = orders.ticket_area_id 
					  AND order_items.order_id = orders.id
			WHERE 
				orders.status != 'failed'
			GROUP BY 
				ticket_area_id, ticket_seat_id
			HAVING 
				COUNT(*) > 1
		) as double_bookings;
	`)

	if err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return &DoubleOrderCheck{
			Total: 0,
		}, nil
	}

	return &raw[0], nil
}
