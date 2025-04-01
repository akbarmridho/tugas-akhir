package main

import (
	"context"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"os"
	"os/signal"
	"sync"
	"time"
	"tugas-akhir/backend/app/processor"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/memcache"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/internal/bookings"
	"tugas-akhir/backend/internal/events"
	"tugas-akhir/backend/internal/orders"
	"tugas-akhir/backend/internal/payments"
	"tugas-akhir/backend/pkg/logger"
)

func main() {
	l := logger.GetInfo()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ctx = logger.WithCtx(ctx, l)

	// todo change the import module so that it match requirements for pgp
	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger.GetInfo()}
		}),
		config.Module,
		memcache.Module,
		postgres.Module,
		bookings.BaseModule,
		events.BaseModule,
		orders.BaseModule,
		payments.BaseModule,
		processor.Module,
		fx.Invoke(func(processor *processor.Processor, c *config.Config) error {
			c.AppVariant = config.AppVariant__PGP

			// todo
			return processor.Run()
		},
		),
	)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		app.Run()
		wg.Done()
	}()

	wg.Wait()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Stop(shutdownCtx); err != nil {
		logger.GetInfo().Info("Application was shutdown ungracefully")
	}
}
