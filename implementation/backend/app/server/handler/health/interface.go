package health

import "github.com/labstack/echo/v4"

type HealthcheckHandler interface {
	Healthcheck(c echo.Context) error
}
