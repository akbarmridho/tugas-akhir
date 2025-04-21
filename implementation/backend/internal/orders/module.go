package orders

import (
	"go.uber.org/fx"
	"tugas-akhir/backend/internal/orders/repository/order"
	"tugas-akhir/backend/internal/orders/service/early_dropper"
	"tugas-akhir/backend/internal/orders/service/pgp_place_order_connector"
	"tugas-akhir/backend/internal/orders/usecase/get_order"
	"tugas-akhir/backend/internal/orders/usecase/place_order"
	"tugas-akhir/backend/internal/orders/usecase/webhook"
)

var BaseModule = fx.Options(
	fx.Provide(fx.Annotate(order.NewPGOrderRepository, fx.As(new(order.OrderRepository)))),
	fx.Provide(fx.Annotate(get_order.NewPGGetOrderUsecase, fx.As(new(get_order.GetOrderUsecase)))),
	fx.Provide(fx.Annotate(place_order.NewBasePlaceOrderUsecase, fx.As(new(place_order.PlaceOrderUsecase)))),
	fx.Provide(fx.Annotate(webhook.NewPGWebhookUsecase, fx.As(new(webhook.WebhookOrderUsecase)))),
)

var FCWorkerModule = fx.Options(
	fx.Provide(fx.Annotate(order.NewPGOrderRepository, fx.As(new(order.OrderRepository)))),
	fx.Provide(fx.Annotate(place_order.NewBasePlaceOrderUsecase, fx.As(new(place_order.PlaceOrderUsecase)))),
)

var FCModule = fx.Options(
	fx.Provide(fx.Annotate(order.NewPGOrderRepository, fx.As(new(order.OrderRepository)))),
	fx.Provide(fx.Annotate(get_order.NewPGGetOrderUsecase, fx.As(new(get_order.GetOrderUsecase)))),
	fx.Provide(fx.Annotate(pgp_place_order_connector.NewFCPlaceOrderConnector,
		fx.OnStart(func(connector *pgp_place_order_connector.FCPlaceOrderConnector) error {
			return connector.Run()
		}),
		fx.OnStop(func(connector *pgp_place_order_connector.FCPlaceOrderConnector) error {
			return connector.Stop()
		}),
	)),
	fx.Provide(fx.Annotate(early_dropper.NewFCEarlyDropper,
		fx.OnStart(func(dropper *early_dropper.EarlyDropper) error {
			return dropper.Run()
		}),
		fx.OnStop(func(dropper *early_dropper.EarlyDropper) error {
			return dropper.Stop()
		}),
	)),
	fx.Provide(fx.Provide(fx.Annotate(
		place_order.NewFCPlaceOrderUsecase,
		fx.As(new(place_order.PlaceOrderUsecase)),
	))),
	fx.Provide(fx.Provide(fx.Annotate(webhook.NewFCWebhookUsecase, fx.As(new(webhook.WebhookOrderUsecase))))),
)
