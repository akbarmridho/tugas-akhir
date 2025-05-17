package sanity

import "tugas-akhir/backend/internal/events/entity"

type AvailabilityCheck struct {
	Count       int
	Available   int
	Unavailable int
}

type DBAvailabilityRow struct {
	Status entity.SeatStatus
	Total  int
}

type DoubleOrderCheck struct {
	Total int
}
