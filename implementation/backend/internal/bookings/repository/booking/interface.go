package booking

import (
	"context"
	"tugas-akhir/backend/internal/bookings/entity"
	entity2 "tugas-akhir/backend/internal/events/entity"
)

type BookingInterface interface {
	Book(ctx context.Context, payload entity.BookingRequestDto) ([]entity2.TicketSeat, error)
}
