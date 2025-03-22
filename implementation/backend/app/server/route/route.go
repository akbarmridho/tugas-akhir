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

func NewRoutes() *Routes {
	rootRoutes := make([]Route, 0)

	return &Routes{
		root: rootRoutes,
	}
}

var Module = fx.Options(
	fx.Provide(NewRoutes),
)
