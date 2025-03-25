package orders

import "github.com/labstack/echo/v4"

type OrderHandler interface {
	PlaceOrder(c echo.Context) error
}
