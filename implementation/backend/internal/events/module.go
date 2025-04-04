package events

import (
	"go.uber.org/fx"
	"tugas-akhir/backend/internal/events/repository/availability"
	"tugas-akhir/backend/internal/events/repository/event"
	"tugas-akhir/backend/internal/events/repository/seat"
	"tugas-akhir/backend/internal/events/usecase"
)

var BaseModule = fx.Options(
	fx.Provide(fx.Annotate(availability.NewPGAvailabilityRepository, fx.As(new(availability.AvailabilityRepository)))),
	fx.Provide(fx.Annotate(event.NewPGEventRepository, fx.As(new(event.EventRepository)))),
	fx.Provide(fx.Annotate(seat.NewPGSeatRepository, fx.As(new(seat.SeatRepository)))),
	fx.Provide(usecase.NewEventAvailabilityUsecase),
)

var PGPModule = fx.Options(
	fx.Provide(fx.Annotate(availability.NewRWAvailabilityRepository, fx.As(new(availability.AvailabilityRepository)))),
	fx.Provide(fx.Annotate(event.NewPGEventRepository, fx.As(new(event.EventRepository)))),
	fx.Provide(fx.Annotate(seat.NewPGSeatRepository, fx.As(new(seat.SeatRepository)))),
	fx.Provide(usecase.NewEventAvailabilityUsecase),
)
