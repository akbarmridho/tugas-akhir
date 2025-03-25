package route

import (
	"github.com/labstack/echo/v4"
	"tugas-akhir/backend/app/server/handler/orders"
)

type WebhookRoute struct {
	webhookHandler orders.OrderHandler
}

func NewWebhookRoute(
	webhookHandler orders.OrderHandler,
) *WebhookRoute {
	return &WebhookRoute{
		webhookHandler: webhookHandler,
	}
}

func (r *WebhookRoute) Setup(engine *echo.Group) {
	group := engine.Group("/webhooks")

	group.POST("/", r.webhookHandler.HandleWebhook)
}
