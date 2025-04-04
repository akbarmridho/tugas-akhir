package availability

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"tugas-akhir/backend/infrastructure/risingwave"
	"tugas-akhir/backend/internal/events/entity"
)

type RWAvailabilityRepository struct {
	rw *risingwave.Risingwave
}

func NewRWAvailabilityRepository(rw *risingwave.Risingwave) *RWAvailabilityRepository {
	return &RWAvailabilityRepository{
		rw: rw,
	}
}

func (r *RWAvailabilityRepository) GetAvailability(ctx context.Context, payload entity.GetAvailabilityDto) ([]entity.AreaAvailability, error) {
	query := `
	SELECT ticket_package_id, ticket_area_id, total_seats, available_seats
	FROM ticket_availability
	WHERE ticket_sale_id = $1
    `

	result := make([]entity.AreaAvailability, 0)

	err := pgxscan.Select(
		ctx,
		r.rw.Pool,
		&result,
		query,
		payload.TicketSaleID,
	)

	if err != nil {
		return nil, err
	}

	// we expect we found data
	if len(result) == 0 {
		return nil, entity.AreaAvailabilityNotFoundError
	}

	return result, nil
}
