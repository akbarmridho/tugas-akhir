package route

import (
	"github.com/labstack/echo/v4"
	"tugas-akhir/backend/app/server/handler/orders"
	"tugas-akhir/backend/app/server/middleware"
)

type OrdersRoute struct {
	authMiddleware *middleware.AuthMiddleware
	orderHandler   orders.OrderHandler
}

func NewOrdersRoute(
	authMiddleware *middleware.AuthMiddleware,
	orderHandler orders.OrderHandler,
) *OrdersRoute {
	return &OrdersRoute{
		authMiddleware: authMiddleware,
		orderHandler:   orderHandler,
	}
}

func (r *OrdersRoute) Setup(engine *echo.Group) {
	group := engine.Group("/orders")
	group.Use(r.authMiddleware.JwtMiddleware)

	group.POST("", r.orderHandler.PlaceOrder)
	group.GET("/:id", r.orderHandler.GetOrder)
	group.GET("/:id/tickets", r.orderHandler.GetIssuedTickets)
}
