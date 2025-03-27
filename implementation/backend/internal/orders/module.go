package orders

import (
	"go.uber.org/fx"
	"tugas-akhir/backend/internal/orders/repository/order"
	"tugas-akhir/backend/internal/orders/usecase/get_order"
	"tugas-akhir/backend/internal/orders/usecase/place_order"
	"tugas-akhir/backend/internal/orders/usecase/webhook"
)

var BaseModule = fx.Options(
	fx.Provide(fx.Provide(fx.Annotate(order.NewPGOrderRepository, fx.As(new(order.OrderRepository))))),
	fx.Provide(fx.Provide(fx.Annotate(get_order.NewPGGetOrderUsecase, fx.As(new(get_order.GetOrderUsecase))))),
	fx.Provide(fx.Provide(fx.Annotate(place_order.NewBasePlaceOrderUsecase, fx.As(new(place_order.PlaceOrderUsecase))))),
	fx.Provide(fx.Provide(fx.Annotate(webhook.NewPGWebhookUsecase, fx.As(new(webhook.WebhookOrderUsecase))))),
)
