package handler

import (
	"go.uber.org/fx"
	"tugas-akhir/backend/app/server/handler/events"
	"tugas-akhir/backend/app/server/handler/health"
	"tugas-akhir/backend/app/server/handler/orders"
)

var BaseModule = fx.Options(
	fx.Provide(events.NewEventHandler),
	fx.Provide(fx.Annotate(health.NewPGHealthcheckHandler, fx.As(new(health.HealthcheckHandler)))),
	fx.Provide(fx.Annotate(orders.NewBaseOrderHandler, fx.As(new(orders.OrderHandler)))),
)
