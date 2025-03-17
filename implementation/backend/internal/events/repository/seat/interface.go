package seat

import (
	"context"
	"tugas-akhir/backend/internal/events/entity"
)

type SeatRepository interface {
	GetSeats(ctx context.Context, payload entity.GetSeatsDto) ([]entity.TicketSeat, error)
}
