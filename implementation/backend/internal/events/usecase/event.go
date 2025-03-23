package usecase

import (
	"context"
	"errors"
	"net/http"
	"tugas-akhir/backend/internal/events/entity"
	"tugas-akhir/backend/internal/events/repository/availability"
	"tugas-akhir/backend/internal/events/repository/event"
	"tugas-akhir/backend/internal/events/repository/seat"
	myerror "tugas-akhir/backend/pkg/error"
)

type EventUsecase struct {
	availabilityRepository availability.AvailabilityRepository
	seatRepository         seat.SeatRepository
	eventRepository        event.EventRepository
}

func NewEventAvailabilityUsecase(
	availabilityRepository availability.AvailabilityRepository,
	seatRepository seat.SeatRepository,
	eventRepository event.EventRepository,
) *EventUsecase {
	return &EventUsecase{
		availabilityRepository: availabilityRepository,
		seatRepository:         seatRepository,
		eventRepository:        eventRepository,
	}
}

func (u *EventUsecase) GetSeats(ctx context.Context, payload entity.GetSeatsDto) ([]entity.TicketSeat, *myerror.HttpError) {
	data, err := u.seatRepository.GetSeats(ctx, payload)

	if err != nil {
		if errors.Is(err, entity.SeatNotFoundError) {
			return nil, &myerror.HttpError{
				Message: err.Error(),
				Code:    http.StatusNotFound,
			}
		} else {
			return nil, &myerror.HttpError{
				Message:      err.Error(),
				ErrorContext: err,
				Code:         http.StatusInternalServerError,
			}
		}
	}

	return data, nil
}

func (u *EventUsecase) GetAvailability(ctx context.Context, payload entity.GetAvailabilityDto) ([]entity.AreaAvailability, *myerror.HttpError) {
	data, err := u.availabilityRepository.GetAvailability(ctx, payload)

	if err != nil {
		if errors.Is(err, entity.AreaAvailabilityNotFoundError) {
			return nil, &myerror.HttpError{
				Message: err.Error(),
				Code:    http.StatusNotFound,
			}
		} else {
			return nil, &myerror.HttpError{
				Message:      err.Error(),
				ErrorContext: err,
				Code:         http.StatusInternalServerError,
			}
		}
	}

	return data, nil
}

func (u *EventUsecase) GetEvent(ctx context.Context, payload entity.GetEventDto) (*entity.Event, *myerror.HttpError) {
	data, err := u.eventRepository.GetEvent(ctx, payload)

	if err != nil {
		if errors.Is(err, entity.EventNotFoundError) {
			return nil, &myerror.HttpError{
				Message: err.Error(),
				Code:    http.StatusNotFound,
			}
		} else {
			return nil, &myerror.HttpError{
				Message:      err.Error(),
				ErrorContext: err,
				Code:         http.StatusInternalServerError,
			}
		}
	}

	return data, nil
}

func (u *EventUsecase) GetEvents(ctx context.Context) ([]entity.Event, *myerror.HttpError) {
	data, err := u.eventRepository.GetEvents(ctx)

	if err != nil {
		return nil, &myerror.HttpError{
			Message:      err.Error(),
			ErrorContext: err,
			Code:         http.StatusInternalServerError,
		}
	}

	return data, nil
}
