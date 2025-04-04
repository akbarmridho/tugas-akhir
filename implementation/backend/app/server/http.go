package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
	"net/http"
	"time"
	"tugas-akhir/backend/app/server/handler/health"
	middleware2 "tugas-akhir/backend/app/server/middleware"
	"tugas-akhir/backend/app/server/route"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/pkg/logger"
)

type Server struct {
	config *config.Config
	engine *echo.Echo
	Port   int
}

func (s *Server) Run() {
	address := fmt.Sprintf(":%d", s.Port)

	// start server
	go func() {
		logger.GetInfo().Sugar().Infof("Server started on port %d", s.Port)
		if err := s.engine.StartTLS(address, s.config.TlsCertPath, s.config.TlsKeyPath); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.engine.Logger.Fatal("Shutting down the server")
		}
	}()
}

func (s *Server) Stop(context context.Context) error {
	return s.engine.Shutdown(context)
}

func NewServer(
	healthHandler health.HealthcheckHandler,
	routes *route.Routes,
	config *config.Config,
	loggerMiddleware *middleware2.LoggerMiddleware,
) *Server {
	engine := echo.New()
	engine.HideBanner = true
	engine.HidePort = true

	engine.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Namespace: "ticket_backend",
		Subsystem: string(config.AppVariant),
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/health"
		},
		LabelFuncs: map[string]echoprometheus.LabelValueFunc{
			"test_scenario":  func(c echo.Context, err error) string { return config.TestScenario },
			"app_variant":    func(c echo.Context, err error) string { return string(config.AppVariant) },
			"kubernetes_pod": func(c echo.Context, err error) string { return config.PodName },
		},
	}))

	engine.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 3 * time.Minute,
	}))

	engine.Use(middleware.RequestID())

	// default error handler
	engine.Use(middleware.Recover())

	engine.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics

	engine.GET("/health", healthHandler.Healthcheck)

	// setup common middleware=
	engine.Use(loggerMiddleware.LoggerMiddleware)

	engine.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			AppVariant string
			Scenario   string
		}{
			AppVariant: string(config.AppVariant),
			Scenario:   config.TestScenario,
		})
	})

	routes.Setup(engine)

	return &Server{
		config: config,
		engine: engine,
		Port:   config.ServerPort,
	}
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewServer,
			fx.OnStop(func(ctx context.Context, server *Server) error {
				return server.Stop(ctx)
			}),
		),
	),
)
