package booked_seats

import (
	"context"
	"tugas-akhir/backend/internal/bookings/entity"
	entity2 "tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/pkg/cursor_iterator"
)

type BookedSeatRepository interface {
	IterSeats(ctx context.Context) ([]entity2.TicketSeat, *cursor_iterator.CursorIterator, error)
	PublishIssuedTickets(ctx context.Context, payload entity.PublishIssuedTicketDto) error
	GetIssuedTickets(ctx context.Context, payload entity.GetIssuedTicketDto) ([]entity.IssuedTicket, error)
}
