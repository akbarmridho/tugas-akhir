package event

import (
	"context"
	"tugas-akhir/backend/internal/events/entity"
)

type EventRepository interface {
	GetEvents(ctx context.Context) ([]entity.Event, error)
	GetEvent(ctx context.Context, payload entity.GetEventDto) (*entity.Event, error)
}
