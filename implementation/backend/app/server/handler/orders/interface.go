package orders

import "github.com/labstack/echo/v4"

const HeaderIdempotencyKey = "Idempotency-Key"

type OrderHandler interface {
	PlaceOrder(c echo.Context) error
	GetOrder(c echo.Context) error
	HandleWebhook(c echo.Context) error
	GetIssuedTickets(c echo.Context) error
}
