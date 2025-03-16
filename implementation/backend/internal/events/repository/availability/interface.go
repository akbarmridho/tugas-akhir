package availability

import (
	"context"
	"tugas-akhir/backend/internal/events/entity"
)

type AvailabilityRepository interface {
	GetAvailability(ctx context.Context, payload entity.GetAvailabilityDto) ([]entity.AreaAvailability, error)
}
