package route

import (
	"github.com/labstack/echo/v4"
	"tugas-akhir/backend/app/server/handler/events"
	"tugas-akhir/backend/app/server/middleware"
)

type EventsRoute struct {
	authMiddleware *middleware.AuthMiddleware
	eventHandler   *events.EventHandler
}

func NewEventsRoute(
	authMiddleware *middleware.AuthMiddleware,
	eventHandler *events.EventHandler,
) *EventsRoute {
	return &EventsRoute{
		authMiddleware: authMiddleware,
		eventHandler:   eventHandler,
	}
}

func (r *EventsRoute) Setup(engine *echo.Group) {
	group := engine.Group("/events")
	group.Use(r.authMiddleware.JwtMiddleware)

	group.GET("/availability/:ticketSaleId", r.eventHandler.GetAvailability)
	group.GET("/seats/:ticketAreaId", r.eventHandler.GetSeats)
	group.GET("/:eventId", r.eventHandler.GetEvent)
	group.GET("/", r.eventHandler.GetEvents)
}
