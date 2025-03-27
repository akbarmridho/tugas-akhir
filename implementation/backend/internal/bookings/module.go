package bookings

import (
	"go.uber.org/fx"
	"tugas-akhir/backend/internal/bookings/repository/booking"
)

var BaseModule = fx.Options(
	fx.Provide(fx.Annotate(booking.NewPGBookingRepository, fx.As(new(booking.BookingRepository)))),
)
