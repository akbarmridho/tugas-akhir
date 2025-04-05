package bookings

import (
	"go.uber.org/fx"
	"tugas-akhir/backend/internal/bookings/repository/booked_seats"
	"tugas-akhir/backend/internal/bookings/repository/booking"
)

var BaseModule = fx.Options(
	fx.Provide(fx.Annotate(booking.NewPGBookingRepository, fx.As(new(booking.BookingRepository)))),
	fx.Provide(fx.Annotate(booked_seats.NewPGBookedSeatRepository, fx.As(new(booked_seats.SeatRepository)))),
)

var ScyllaModule = fx.Options(
	fx.Provide(fx.Annotate(booking.NewScyllaBookingRepository, fx.As(new(booking.BookingRepository)))),
	fx.Provide(fx.Annotate(booked_seats.NewPGBookedSeatRepository, fx.As(new(booked_seats.SeatRepository)))),
)
