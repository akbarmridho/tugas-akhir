package route

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

type Route interface {
	Setup(engine *echo.Group)
}

type Routes struct {
	root []Route
}

func (r Routes) Setup(engine *echo.Echo) {
	rootGroup := engine.Group("")

	for _, route := range r.root {
		route.Setup(rootGroup)
	}
}

func NewRoutes(
	eventsRoute *EventsRoute,
	ordersRoute *OrdersRoute,
	webhookRoute *WebhookRoute,
) *Routes {
	rootRoutes := []Route{
		eventsRoute,
		ordersRoute,
		webhookRoute,
	}

	return &Routes{
		root: rootRoutes,
	}
}

var Module = fx.Options(
	fx.Provide(NewEventsRoute),
	fx.Provide(NewOrdersRoute),
	fx.Provide(NewRoutes),
	fx.Provide(NewWebhookRoute),
)
