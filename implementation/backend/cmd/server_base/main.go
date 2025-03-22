package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"
	"tugas-akhir/backend/app/server"
	"tugas-akhir/backend/app/server/middleware"
	"tugas-akhir/backend/app/server/route"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/pkg/logger"
	myvalidator "tugas-akhir/backend/pkg/validator"

	_ "go.uber.org/automaxprocs"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func RunServer(server *server.Server, c *config.Config) {
	c.AppVariant = config.AppVariant__Radar
	server.Run()
}

func main() {
	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger.GetInfo()}
		}),
		fx.Provide(myvalidator.NewTranslastedValidator),
		config.Module,
		middleware.Module,
		route.Module,
		server.Module,
		fx.Invoke(RunServer),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		app.Run()
		wg.Done()
	}()

	wg.Wait()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		logger.GetInfo().Info("Application was shutdown ungracefully")
	}
}
