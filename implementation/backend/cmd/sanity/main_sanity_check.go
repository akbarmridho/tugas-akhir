package main

import (
	"context"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"os"
	"os/signal"
	"sync"
	"time"
	"tugas-akhir/backend/app/processor"
	"tugas-akhir/backend/app/sanity"
	"tugas-akhir/backend/infrastructure/config"
	"tugas-akhir/backend/infrastructure/memcache"
	"tugas-akhir/backend/infrastructure/postgres"
	"tugas-akhir/backend/infrastructure/redis"
	"tugas-akhir/backend/pkg/logger"
)

func main() {
	l := logger.GetInfo()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ctx = logger.WithCtx(ctx, l)

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger.GetInfo()}
		}),
		config.Module,
		memcache.Module,
		postgres.Module,
		redis.Module,
		fx.Provide(fx.Annotate(
			processor.NewMetricsServer,
			fx.OnStop(func(ctx context.Context, metricsServer *processor.MetricsServer) error {
				return metricsServer.Stop(ctx)
			}),
		)),
		fx.Provide(fx.Annotate(
			sanity.NewSanityCheck,
			fx.OnStop(func(sanityCheck *sanity.SanityCheck) error {
				sanityCheck.Stop()
				return nil
			}),
		)),
		fx.Invoke(func(sanityCheck *sanity.SanityCheck, metricsServer *processor.MetricsServer, c *config.Config) error {
			sanityCheck.Run(ctx)

			if err := metricsServer.Run(ctx); err != nil {
				return err
			}

			return nil
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
