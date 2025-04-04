package availability

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/events/entity"
)

type PGAvailabilityRepository struct {
	pg *postgres.Postgres
}

func NewPGAvailabilityRepository(pg *postgres.Postgres) *PGAvailabilityRepository {
	return &PGAvailabilityRepository{
		pg: pg,
	}
}

func (r *PGAvailabilityRepository) GetAvailability(ctx context.Context, payload entity.GetAvailabilityDto) ([]entity.AreaAvailability, error) {
	query := `
	SELECT 
		tp.id AS ticket_package_id,
		ta.id AS ticket_area_id,
		COUNT(ts.id) AS total_seats,
		COUNT(CASE WHEN ts.status = 'available' THEN 1 END) AS available_seats
	FROM 
		ticket_packages tp
	INNER JOIN 
		ticket_areas ta ON ta.ticket_package_id = tp.id
	INNER JOIN 
		ticket_seats ts ON ts.ticket_area_id = ta.id
	WHERE 
		tp.ticket_sale_id = $1
	GROUP BY 
		tp.id, ta.id
    `

	result := make([]entity.AreaAvailability, 0)

	err := pgxscan.Select(
		ctx,
		r.pg.GetExecutor(ctx),
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
